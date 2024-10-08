package utils

import (
	"net/http"

	"github.com/realclientip/realclientip-go"
)

func GetIP(r *http.Request) string {
	strat, err := realclientip.NewRightmostNonPrivateStrategy("X-Forwarded-For")
	if err != nil {
		return "0.0.0.0"
	}

	return strat.ClientIP(r.Header, r.RemoteAddr)
}

func GetUserAgent(r *http.Request) string {
	return r.UserAgent()
}

func SetResponseHeaders(w http.ResponseWriter) {
	w.Header().Add("Link", "rel=preload; </static/css/reset.css>; as=style")
	w.Header().Add("Link", "rel=preload; </static/css/app.css>; as=style")
	w.Header().Add("Link", "rel=preload; </static/css/table.css>; as=style")
	w.Header().Add("Link", "rel=preload; </static/css/dialog.css>; as=style")
	w.Header().Add("Link", "rel=preload; </static/css/icon.css>; as=style")
	w.Header().Add("Link", "rel=preload; </static/js/app.js>; as=script")
}
