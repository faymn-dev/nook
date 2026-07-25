package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("public")))

	server := http.Server{
		Addr:    ":8888",
		Handler: mux,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("Failed to start server.")
	}
}
