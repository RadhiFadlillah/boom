package webserver

import (
	"net/http"
	fp "path/filepath"
	"strings"

	"github.com/go-boom/boom/internal/build"
	"github.com/julienschmidt/httprouter"
)

// Handler is handler for serving the web interface.
type Handler struct {
	build.Worker
}

func newHandler(rootDir string) (Handler, error) {
	worker, err := build.NewWorker(rootDir, false, true)
	if err != nil {
		return Handler{}, err
	}

	return Handler{Worker: worker}, nil
}

func (hdl *Handler) serveSite(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse URL
	urlPath := strings.Trim(r.URL.Path, "/")
	pathSegments := strings.Split(urlPath, "/")

	// Make sure to disable cache
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// If it's for assets and themes, just serve it directly
	if len(pathSegments) > 0 && (pathSegments[0] == "assets" || pathSegments[0] == "themes") {
		staticPath := fp.Join(hdl.Worker.RootDir, urlPath)
		http.ServeFile(w, r, staticPath)
		return
	}

	// If not, it must be content that need to be build
	w.Header().Set("Content-Type", "text/html")
	_, err := hdl.Build(urlPath, w)
	panicError(err)
}
