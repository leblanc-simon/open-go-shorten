package utils

import (
	"net/http"
)

func SetResponseHeaders(w http.ResponseWriter) {
	w.Header().Add("Link", "rel=preload; </static/css/reset.css>; as=style")
	w.Header().Add("Link", "rel=preload; </static/css/app.css>; as=style")
	w.Header().Add("Link", "rel=preload; </static/css/table.css>; as=style")
	w.Header().Add("Link", "rel=preload; </static/css/dialog.css>; as=style")
	w.Header().Add("Link", "rel=preload; </static/css/icon.css>; as=style")
	w.Header().Add("Link", "rel=preload; </static/js/app.js>; as=script")
}
