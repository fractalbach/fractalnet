package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	//"net"
	"net/http"
	"strings"
	"time"
)

var addr = flag.String("a", "localhost:8080", "http service address")
var example = &Game{
	GameId:      1337,
	Description: "Some Great Game.",
	special:     "omgz",
}

// Game is an example of one of the highest level objects.
// It is made of up of different worlds within a game.
type Game struct {
	GameId      int
	Description string
	special     string
}

func (g *Game) fetchInfo() []byte {
	out, err := json.Marshal(g)
	if err != nil {
		log.Println(err)
		return []byte("There was an error converting game data into json.")
	}
	return out
}

/*
func (g *Game) MarshalJSON() {

}*/

// World is an object that players can actively join.
// World can be made up of zones
type World struct {
	worldId     int
	description string
}

// Zone is a specific location chunk within a World.
type Zone struct {
	zoneId      int
	description string
}

type Fetchable interface {
	fetchInfo() []byte
}

func Display(w http.ResponseWriter, f Fetchable) {
	fmt.Fprintln(w, string(f.fetchInfo()))
}

type Page struct {
	Title string
	Body  []byte
}

func main() {
	log.Println("Starting FractalNet")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/", serveAPI)
	mux.HandleFunc("/", serveIndex)

	s := &http.Server{
		Addr:           (*addr),
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Listening and Serving on http://%v/", (*addr))
	log.Fatal(s.ListenAndServe())
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	logRequest(r)

	/*
	   if r.Method != "GET" {
	       http.Error(w, "Method not allowed", 405)
	   }
	*/
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	fmt.Fprintln(w, "Welcome to Fractal Net Homepage!")

	/*
		    r.ParseForm()
			fmt.Fprintln(w, "scheme", r.URL.Scheme)
			for k, v := range r.Form {
				fmt.Fprintln(w, k, v)
			}
	*/
}

func serveModifyUser(w http.ResponseWriter, r *http.Request) {

}

// serveAPI is the main handler that deals with all of the incoming requests
// to display API information.
func serveAPI(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	switch strings.ToLower(r.URL.Path) {
	case "/api/":
		myUrl := "http://" + r.Host + "/"
		fmt.Fprintln(w, "Welcome to the Api, try going to "+myUrl+"api/games/")

	case "/api/games/":
		listGames(w, r)

	case "/api/players/":
		listPlayers(w, r)

	default:
		http.NotFound(w, r)
	}
}

func listGames(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "This will display a list of games.")
	f := example
	Display(w, f)
}

func listPlayers(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "This displays a list of players.")
}

// logRequest prints out a useful message to the command line log,
// displaying information about the request that was just made to the server.
func logRequest(r *http.Request) {
	log.Printf("(%v) %v %v %v", r.RemoteAddr, r.Proto, r.Method, r.URL)
}
