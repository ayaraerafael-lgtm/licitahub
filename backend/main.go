package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(healthResponse{
			Status:  "ok",
			Service: "licitahub-api",
		})
	})

	log.Println("LicitaHub API listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
