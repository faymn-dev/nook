package main

import (
	"context"
	"net/http"

	"github.com/urfave/cli/v3"
)

func servePublicAction(context.Context, *cli.Command) error {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("public")))

	server := http.Server{
		Addr:    ":8888",
		Handler: mux,
	}

	return server.ListenAndServe()
}
