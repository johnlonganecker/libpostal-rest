package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

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

	host := os.GetEnv("LISTEN_HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	port := os.GetEnv("LISTEN_PORT")
	if port == "" {
		port = "5000"
	}
	certFile := os.GetEnv("SSL_CERT_FILE")
	keyFile := os.GetEnv("SSL_KEY_FILE")
	listenSpec := fmt.Sprintf("%s:%s", host, port)

	if certFile != "" && keyFile != "" {
		fmt.Printf("listening on https://%s\n", listenSpec)
		http.ListenAndServeTLS(listenSpec, certFile, keyFile, r)
	} else {
		fmt.Printf("listening on http://%s\n", listenSpec)
		http.ListenAndServe(listenSpec, r)
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
