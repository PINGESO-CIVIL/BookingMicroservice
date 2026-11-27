package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

func main() {
	const address = ":8080"
	if len(os.Args) == 2 && os.Args[1] == "healthcheck" {
		client := &http.Client{Timeout: 3 * time.Second}
		response, err := client.Get("http://127.0.0.1:8080/healthz")
		if err != nil {
			log.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		return
	}
	backend, err := url.Parse(os.Getenv("BOOKING_API_URL"))
	if err != nil || backend.Host == "" || (backend.Scheme != "http" && backend.Scheme != "https") {
		log.Fatal("BOOKING_API_URL debe ser una URL HTTP válida")
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = 15 * time.Second
	proxy := &httputil.ReverseProxy{
		Rewrite: func(request *httputil.ProxyRequest) {
			request.SetURL(backend)
			request.SetXForwarded()
		},
		Transport: transport,
	}
	root := http.Dir("/app/dist")
	static := http.FileServer(root)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	mux.Handle("/api/", proxy)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}
		file, err := root.Open(r.URL.Path)
		if err != nil {
			// Navigation paths resolve to the app. Missing static assets retain 404.
			if r.Header.Get("Sec-Fetch-Mode") == "navigate" {
				http.ServeFile(w, r, "/app/dist/index.html")
				return
			}
		} else {
			file.Close()
		}
		static.ServeHTTP(w, r)
	})
	server := &http.Server{Addr: address, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(server.ListenAndServe())
}
