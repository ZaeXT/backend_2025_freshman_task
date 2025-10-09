package handlers

import (
	"net/http"
)

// ServeHTML 提供HTML页面
func ServeHTML(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "index.html")
}
