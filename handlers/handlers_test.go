package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/golang-jwt/jwt/v5"
	"open-go-shorten.eu/config"
	"open-go-shorten.eu/middleware"
	"open-go-shorten.eu/models"
	"open-go-shorten.eu/utils"
)

const testSecret = "test-secret"

func setup(t *testing.T) (*miniredis.Miniredis, http.Handler) {
	t.Helper()

	mr := miniredis.RunT(t)

	cfg := &config.Config{}
	cfg.Auth.Username = "admin"
	cfg.Auth.Password = config.DefaultPasswordHash
	cfg.Auth.JwtSecret = testSecret
	cfg.Database.Host = mr.Host()
	cfg.Database.Port, _ = strconv.Atoi(mr.Port())
	cfg.Database.Prefix = "ogs-"
	cfg.Server.PublicURL = "https://s.example.com/"

	utils.InitRedis(cfg)
	middleware.InitJwt(cfg)
	InitAuth(cfg)
	if err := InitQRCode(cfg); err != nil {
		t.Fatal(err)
	}
	InitClientIP(func(r *http.Request) string { return "192.0.2.1" })

	mux := http.NewServeMux()
	mux.HandleFunc("POST /login", Login)
	mux.Handle("POST /shorten", middleware.JWTMiddleware(http.HandlerFunc(ShortenURL)))
	mux.Handle("GET /urls", middleware.JWTMiddleware(http.HandlerFunc(GetURLs)))
	mux.Handle("DELETE /{shortURL}", middleware.JWTMiddleware(http.HandlerFunc(DeleteUrl)))
	mux.HandleFunc("GET /qrcode/{shortURL}/{format}", QRCode)
	mux.HandleFunc("GET /{shortURL}", RedirectURL)

	return mr, mux
}

func validToken(t *testing.T) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "admin",
		"exp":      time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func do(h http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func shorten(t *testing.T, h http.Handler, body string) string {
	t.Helper()
	rec := do(h, "POST", "/shorten", body, validToken(t))
	if rec.Code != http.StatusOK {
		t.Fatalf("shorten: got %d: %s", rec.Code, rec.Body)
	}
	var res map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	return res["shortURL"]
}

func TestLogin(t *testing.T) {
	_, h := setup(t)

	cases := []struct {
		body string
		want int
	}{
		{`{"username":"admin","password":"open-go-shorten"}`, http.StatusOK},
		{`{"username":"admin","password":"wrong"}`, http.StatusUnauthorized},
		{`{"username":"other","password":"open-go-shorten"}`, http.StatusUnauthorized},
		{`not json`, http.StatusBadRequest},
	}
	for _, c := range cases {
		if rec := do(h, "POST", "/login", c.body, ""); rec.Code != c.want {
			t.Errorf("%s: got %d, want %d", c.body, rec.Code, c.want)
		}
	}
}

func TestJWTMiddleware(t *testing.T) {
	_, h := setup(t)

	expired, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(-time.Hour).Unix(),
	}).SignedString([]byte(testSecret))
	noExp, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{}).SignedString([]byte(testSecret))
	otherAlg, _ := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(testSecret))

	for name, header := range map[string]string{
		"missing":   "",
		"no bearer": "garbage",
		"empty":     "Bearer ",
		"expired":   "Bearer " + expired,
		"no exp":    "Bearer " + noExp,
		"other alg": "Bearer " + otherAlg,
	} {
		req := httptest.NewRequest("GET", "/urls", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s: got %d, want 401", name, rec.Code)
		}
	}

	if rec := do(h, "GET", "/urls", "", validToken(t)); rec.Code != http.StatusOK {
		t.Errorf("valid token: got %d", rec.Code)
	}
}

func TestShortenValidation(t *testing.T) {
	_, h := setup(t)

	past := time.Now().Add(-time.Hour).Format(time.RFC3339)
	for _, body := range []string{
		`{"url":""}`,
		`{"url":"javascript:alert(1)"}`,
		`{"url":"/relative"}`,
		`{"url":"https://example.com","expiration":"` + past + `"}`,
		`{`,
	} {
		if rec := do(h, "POST", "/shorten", body, validToken(t)); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: got %d, want 400", body, rec.Code)
		}
	}
}

func TestShortenRedirectAndStats(t *testing.T) {
	mr, h := setup(t)

	short := shorten(t, h, `{"url":"https://example.com/page"}`)
	if again := shorten(t, h, `{"url":"https://example.com/page"}`); again != short {
		t.Errorf("same URL gave %q then %q", short, again)
	}

	for i := 0; i < 2; i++ {
		rec := do(h, "GET", "/"+short, "", "")
		if rec.Code != http.StatusFound {
			t.Fatalf("redirect: got %d", rec.Code)
		}
		if loc := rec.Header().Get("Location"); loc != "https://example.com/page" {
			t.Fatalf("redirect to %q", loc)
		}
	}

	expiration := time.Now().Add(48 * time.Hour).UTC().Truncate(time.Second)
	expiring := shorten(t, h, `{"url":"https://example.com/other","expiration":"`+expiration.Format(time.RFC3339)+`"}`)
	if ttl := mr.TTL("ogs-" + expiring); ttl <= 0 {
		t.Errorf("expiring URL has TTL %v", ttl)
	}

	// A key outside of the prefix must be ignored.
	mr.Set("other-app-key", "value")

	rec := do(h, "GET", "/urls", "", validToken(t))
	var history []models.History
	if err := json.NewDecoder(rec.Body).Decode(&history); err != nil {
		t.Fatal(err)
	}
	if len(history) != 2 {
		t.Fatalf("got %d URLs, want 2: %+v", len(history), history)
	}
	for _, item := range history {
		switch item.Shorten {
		case short:
			if item.URLData.Expiration != nil {
				t.Errorf("permanent URL has expiration %v", item.URLData.Expiration)
			}
			if item.Stats.VisitorsCount != 2 || item.Stats.UniqueVisitorsCount != 1 {
				t.Errorf("stats: %+v", item.Stats)
			}
		case expiring:
			if item.URLData.Expiration == nil {
				t.Error("expiring URL has no expiration")
			}
		default:
			t.Errorf("unexpected short URL %q", item.Shorten)
		}
	}
}

func TestEmptyList(t *testing.T) {
	_, h := setup(t)

	rec := do(h, "GET", "/urls", "", validToken(t))
	if body := strings.TrimSpace(rec.Body.String()); body != "[]" {
		t.Errorf("got %s, want []", body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type %q", ct)
	}
}

func TestCollisionDoesNotOverwrite(t *testing.T) {
	mr, h := setup(t)

	code, err := utils.GenerateShortURL("https://example.com/new", 0)
	if err != nil {
		t.Fatal(err)
	}
	mr.Set("ogs-"+code, "https://example.com/existing")

	short := shorten(t, h, `{"url":"https://example.com/new"}`)
	if short == code {
		t.Fatal("colliding short URL was reused")
	}
	if got, _ := mr.Get("ogs-" + code); got != "https://example.com/existing" {
		t.Errorf("existing URL overwritten by %q", got)
	}
}

func TestDelete(t *testing.T) {
	_, h := setup(t)

	short := shorten(t, h, `{"url":"https://example.com/delete"}`)
	do(h, "GET", "/"+short, "", "")

	if rec := do(h, "DELETE", "/"+short, "", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("delete without token: got %d", rec.Code)
	}
	if rec := do(h, "DELETE", "/"+short, "", validToken(t)); rec.Code != http.StatusOK {
		t.Errorf("delete: got %d", rec.Code)
	}
	if rec := do(h, "GET", "/"+short, "", ""); rec.Code != http.StatusNotFound {
		t.Errorf("redirect after delete: got %d", rec.Code)
	}
}

func TestQRCode(t *testing.T) {
	_, h := setup(t)

	short := shorten(t, h, `{"url":"https://example.com/qr"}`)

	rec := do(h, "GET", "/qrcode/"+short+"/png?download=1", "", "")
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("png: got %d %s", rec.Code, rec.Header().Get("Content-Type"))
	}
	if cd := rec.Header().Get("Content-Disposition"); cd != `attachment; filename="`+short+`.png"` {
		t.Errorf("Content-Disposition %q", cd)
	}

	rec = do(h, "GET", "/qrcode/"+short+"/svg", "", "")
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/svg+xml" {
		t.Fatalf("svg: got %d %s", rec.Code, rec.Header().Get("Content-Type"))
	}
	if rec.Header().Get("Content-Disposition") != "" {
		t.Error("svg should be inline without ?download")
	}

	for _, path := range []string{
		"/qrcode/" + short + "/gif",
		"/qrcode/unknown1/png",
		"/qrcode/visits-" + short + "/png",
	} {
		if rec := do(h, "GET", path, "", ""); rec.Code != http.StatusNotFound {
			t.Errorf("%s: got %d, want 404", path, rec.Code)
		}
	}
}

func TestBaseURL(t *testing.T) {
	publicURL = ""
	t.Cleanup(func() { publicURL = "" })

	req := httptest.NewRequest("GET", "http://s.example.org/qrcode/x/svg", nil)
	if got := baseURL(req); got != "http://s.example.org" {
		t.Errorf("got %q", got)
	}
	req.Header.Set("X-Forwarded-Proto", "https")
	if got := baseURL(req); got != "https://s.example.org" {
		t.Errorf("got %q", got)
	}

	publicURL = "https://short.example"
	if got := baseURL(req); got != "https://short.example" {
		t.Errorf("got %q", got)
	}
}
