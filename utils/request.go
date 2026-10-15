package utils

import (
	"net/http"
)

func SetResponseHeaders(w http.ResponseWriter) {
	w.Header().Add("Link", "</static/css/reset.css>; rel=preload; as=style")
	w.Header().Add("Link", "</static/css/app.css>; rel=preload; as=style")
	w.Header().Add("Link", "</static/css/table.css>; rel=preload; as=style")
	w.Header().Add("Link", "</static/css/dialog.css>; rel=preload; as=style")
	w.Header().Add("Link", "</static/css/icon.css>; rel=preload; as=style")
	w.Header().Add("Link", "</static/js/app.js>; rel=preload; as=script")
}
