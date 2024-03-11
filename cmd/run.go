package main

import (
	"io/fs"
	"net/http"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/rubbs/mytischtennis-viewer/assets"
	"github.com/rubbs/mytischtennis-viewer/parser"
	log "github.com/sirupsen/logrus"
)

type route struct {
	method  string
	regex   *regexp.Regexp
	handler http.Handler
}

func newRoute(method, pattern string, handler http.Handler) route {
	return route{method, regexp.MustCompile("^" + pattern + "$"), handler}
}


func main() {

	log.Info("starting new server instance")
	r := http.NewServeMux()

	content, err := fs.Sub(fs.FS(assets.Files), ".")
	if err != nil {
		log.WithError(err).Panic("assets not found")
	}

	// Serve static files
	r.Handle("/css/", http.StripPrefix("/", http.FileServer(http.FS(content))))
	r.Handle("/js/", http.StripPrefix("/", http.FileServer(http.FS(content))))
	r.Handle("/img/", http.StripPrefix("/", http.FileServer(http.FS(content))))
	r.Handle("/fonts/", http.StripPrefix("/", http.FileServer(http.FS(content))))

	// create routes
	routes := []route{}
	routes = append(routes, newRoute("GET", "/", &mainHandler{}))

	// custom handler for all other requests
	r.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info(r)
		var allow []string
		for _, route := range routes {
			matches := route.regex.FindStringSubmatch(r.URL.Path)
			if len(matches) > 0 {
				if r.Method != route.method {
					allow = append(allow, route.method)
					continue
				}
				route.handler.ServeHTTP(w, r)
				return
			}
		}
		if len(allow) > 0 {
			w.Header().Set("Allow", strings.Join(allow, ", "))
			http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
			return
		}
		http.NotFound(w, r)
	}))

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	// Start HTTP server.
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}

type mainHandler struct{

}

func (mh *mainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	game := parser.Parse("https://www.mytischtennis.de/clicktt/TTBW/23-24/ligen/Bezirksklasse-Nord-Herren/gruppe/445912/spielbericht/14553473/SF-Gechingen-vs-VfL-Nagold/")
	// log.Info(game)

	// Note the call to ParseFS instead of Parse
	t, err := template.ParseFS(assets.Files, "templates/main.html.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	t.Execute(w, game)
}
