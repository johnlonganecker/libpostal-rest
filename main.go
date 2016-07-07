package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	expand "github.com/openvenues/gopostal/expand"
	parser "github.com/openvenues/gopostal/parser"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/expand", ExpandHandler)
	r.HandleFunc("/parser", ParserHandler)
	r.HandleFunc("/", HomeHandler)
	fmt.Println("listening on port 080")
	http.ListenAndServe(":8080", r)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Gorilla!\n"))
}

func ExpandHandler(w http.ResponseWriter, r *http.Request) {
	expansions := expand.ExpandAddress("Quatre-vingt-douze Ave des Ave des Champs-Élysées")

	for i := 0; i < len(expansions); i++ {
		//fmt.Println(expansions[i])
		w.Write([]byte(expansions[i]))
		w.Write([]byte("\n"))
	}
}

func ParserHandler(w http.ResponseWriter, r *http.Request) {
	parsed := parser.ParseAddress("781 Franklin Ave Crown Heights Brooklyn NY 11216 USA")
	fmt.Println(parsed)
	//w.Write([]byte(parsed))
	w.Write([]byte("\n"))
}
