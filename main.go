package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	expand "github.com/openvenues/gopostal/expand"
	parser "github.com/openvenues/gopostal/parser"
)

type Request struct {
	Query string `json:"query"`
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/expand", ExpandHandler).Methods("POST")
	r.HandleFunc("/parser", ParserHandler).Methods("POST")

	certFile := flag.String("certfile", "", "SSL Cert file")
	keyFile := flag.String("keyfile", "", "SSL Key file")
	host := flag.String("listen-host", "0.0.0.0", "Listen host")
	port := flag.Int("listen-port", 8080, "Listen port")
	listenSpec := fmt.Sprintf("%s:%d", host, port)

	fmt.Printf("listening on port %d", port)
	if certFile != "" && keyFile != "" {
		http.ListenAndServe(listenSpec, r)
	} else {
		http.ListenAndServeTLS(certFile, keyFile, r)
	}
}

func ExpandHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req Request

	q, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(q, &req)

	expansions := expand.ExpandAddress(req.Query)

	expansionThing, _ := json.Marshal(expansions)
	w.Write(expansionThing)
}

func ParserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req Request

	q, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(q, &req)

	parsed := parser.ParseAddress(req.Query)
	parseThing, _ := json.Marshal(parsed)
	w.Write(parseThing)
}
