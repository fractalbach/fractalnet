package main

import (
	"flag"
	"log"
	"net/http"
	//"os"
	"time"

	"github.com/fractalbach/fractalnet/wschat"
)

var addr = flag.String("a", "localhost:8080", "http service address")

func main() {
	log.Println("Starting up Fractal Game Net...")
	flag.Parse()
	/*
		addr := "localhost:8080"

		arguments := os.Args[1:]

		if len(arguments) >= 1 {
			addr = os.Args[1]
		}*/

	log.Println("Starting Websocket Hub...")
	hub := wschat.NewHub()
	go hub.Run()

	/*
		Create a Custom Server Multiplexer

		"servemux is an http request multiplexer. it matches the url of each
		incoming request against a list of registered patterns and calls the
		handler for the pattern that most closely matches the url."
			https://golang.org/pkg/net/http/#ServeMux
	*/
	mux := http.NewServeMux()
	mux.HandleFunc("/", serveHome)
	mux.HandleFunc("/favicon.ico", faviconHandler)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wschat.ServeWs(hub, w, r)
	})

	// Define parameters for running a custom HTTP server
	s := &http.Server{
		Addr:           *addr,
		Handler:        mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Println("Listening and Serving on ", (*addr))
	log.Fatal(s.ListenAndServe())
}

/*
serveHome controls which files are accessible on the server based on how
the server responds to requests for those files.
*/
func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	http.ServeFile(w, r, "war.html")
	return
	/*
		switch r.URL.Path {
		case "/":
			//http.ServeFile(w, r, "index.html")
			http.ServeFile(w, r, "war.html")
			return

		case "/gamechat.html":
			http.ServeFile(w, r, "gamechat.html")
			return

		case "/war.html":
			http.ServeFile(w, r, "war.html")
			return

		default:
			http.Error(w, "Not found", 404)
			return
		}
	*/
	http.Error(w, "Bad Request.", 400)
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "favicon.ico")
}
