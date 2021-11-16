package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/mux"
	expand "github.com/openvenues/gopostal/expand"
	parser "github.com/openvenues/gopostal/parser"
)

type Request struct {
	Query string `json:"address"`
}

type Response struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type OutputResponse struct {
	Street   string
	City     string
	State    string
	Postcode string
	Country  string
}

type MultiOutputResponse struct {
	Outputs []OutputResponse
}

func (output *MultiOutputResponse) AddItem(item OutputResponse) []OutputResponse {
	output.Outputs = append(output.Outputs, item)
	return output.Outputs
}

func main() {
	host := os.Getenv("LISTEN_HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	port := os.Getenv("LISTEN_PORT")
	if port == "" {
		port = "8081"
	}
	listenSpec := fmt.Sprintf("%s:%s", host, port)

	certFile := os.Getenv("SSL_CERT_FILE")
	keyFile := os.Getenv("SSL_KEY_FILE")

	router := mux.NewRouter()
	router.HandleFunc("/health", HealthHandler).Methods("GET")
	router.HandleFunc("/expand", ExpandHandler).Methods("POST")
	router.HandleFunc("/parser", ParserHandler).Methods("POST")
	router.HandleFunc("/multi-parser", MultiParserHandler).Methods("POST")

	s := &http.Server{Addr: listenSpec, Handler: router}
	go func() {
		if certFile != "" && keyFile != "" {
			fmt.Printf("listening on https://%s\n", listenSpec)
			s.ListenAndServeTLS(certFile, keyFile)
		} else {
			fmt.Printf("listening on http://%s\n", listenSpec)
			s.ListenAndServe()
		}
	}()

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	<-stop
	fmt.Println("\nShutting down the server...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	s.Shutdown(ctx)
	fmt.Println("Server stopped")
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
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

func MultiParserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req []Request
	var res []Response

	q, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(q, &req)

	final_outputs := MultiOutputResponse{}

	for r := range req {
		parsed := parser.ParseAddress(req[r].Query)
		parseThing, _ := json.Marshal(parsed)
		json.Unmarshal(parseThing, &res)

		var Street strings.Builder
		var City strings.Builder
		var State strings.Builder
		var Postcode strings.Builder
		var Country strings.Builder
		var Suburb strings.Builder
		var District strings.Builder

		for l := range res {
			res_label := res[l].Label
			res_value := res[l].Value

			switch res_label {
			case "house":
				Street.WriteString(res_value)
				Street.WriteString(" ")
			case "house_number":
				Street.WriteString(res_value)
				Street.WriteString(" ")
			case "unit":
				Street.WriteString(res_value)
				Street.WriteString(" ")
			case "level":
				Street.WriteString(res_value)
				Street.WriteString(" ")
			case "road":
				Street.WriteString(res_value)
				Street.WriteString(" ")
			case "po_box":
				Street.WriteString(res_value)
				Street.WriteString(" ")
			case "entrance":
				Street.WriteString(res_value)
				Street.WriteString(" ")
			case "near":
				Street.WriteString(res_value)
				Street.WriteString(" ")
			case "island":
				Street.WriteString(res_value)
				Street.WriteString(" ")
			case "city_district":
				Street.WriteString(res_value)
				Street.WriteString(" ")
			case "state_district":
				District.WriteString(res_value)
				District.WriteString(" ")
			case "suburb":
				Suburb.WriteString(res_value)
				Suburb.WriteString(" ")
			case "country":
				Country.WriteString(res_value)
			case "postcode":
				Postcode.WriteString(res_value)
			case "city":
				City.WriteString(res_value)
			case "state":
				State.WriteString(res_value)
			}
		}
		if len(City.String()) > 0 {
			Street.WriteString(Suburb.String())
			Street.WriteString(District.String())
		} else {
			if len(Suburb.String()) > 0 {
				if len(District.String()) > 0 {
					City.WriteString(Suburb.String())
					Street.WriteString(District.String())
				} else {
					City.WriteString(Suburb.String())
				}
			} else {
				if len(District.String()) > 0 {
					City.WriteString(District.String())
				}
			}

		}
		out := OutputResponse{Street: strings.TrimSpace(Street.String()),
			City:     City.String(),
			State:    State.String(),
			Postcode: Postcode.String(),
			Country:  Country.String()}

		final_outputs.AddItem(out)
	}
	outputparsed, _ := json.Marshal(final_outputs)
	w.Write(outputparsed)
}
