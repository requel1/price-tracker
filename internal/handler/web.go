package handler

import (
	"net/http"
	"os"
	"path/filepath"
)

type WebHandler struct {
	webDir string
}

func NewWebHandler(webDir string) *WebHandler {
	return &WebHandler{webDir: webDir}
}

func (h *WebHandler) Index(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(h.webDir, "index.html")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, path)
}
