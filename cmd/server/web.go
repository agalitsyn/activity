package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"github.com/Masterminds/sprig"
	"github.com/go-pkgz/routegroup"

	"github.com/agalitsyn/activity/cmd/server/controller"
	"github.com/agalitsyn/activity/version"
)

func NewRouter(
	pageCtrl *controller.PageController,
	wsCtrl *controller.WebSocketController,
) *routegroup.Bundle {
	router := routegroup.New(http.NewServeMux())

	router.Use(RequestLogger, RecovererMiddleware)

	router.Handle("/static/", FileServerHandlerFunc(EmbedFiles, "static"))
	router.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("User-agent: *\nDisallow: /"))
	})

	// Stub browser requests on favicon
	router.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	router.HandleFunc("/agent", wsCtrl.HandleAgent)
	router.HandleFunc("/browser", wsCtrl.HandleBrowser)
	router.HandleFunc("/", pageCtrl.HomePage)
	return router
}

func templateFuncs() template.FuncMap {
	funcs := sprig.FuncMap()
	funcs["printVersion"] = printVersion
	return funcs
}

func printVersion() string {
	return version.String()
}

func FileServerHandlerFunc(embedFiles embed.FS, staticFolder string) http.HandlerFunc {
	staticFS, err := fs.Sub(embedFiles, staticFolder) // error is always nil
	if err != nil {
		panic(err) // should never happen we load from embedded FS
	}
	return func(w http.ResponseWriter, r *http.Request) {
		webFS := http.StripPrefix(fmt.Sprintf("/%s/", staticFolder), http.FileServer(http.FS(staticFS)))
		webFS.ServeHTTP(w, r)
	}
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("INFO", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func RecovererMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println("ERROR", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
