package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	command := "build"
	if len(os.Args) >= 2 {
		command = os.Args[1]
	}

	switch command {
	case "build":
		panic("not implemented yet")
	case "serve":
		servePublic()
	default:
		fmt.Printf("invalid command %q\n", command)
		os.Exit(1)
	}
}

func servePublic() {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("public")))

	server := http.Server{
		Addr:    ":8888",
		Handler: mux,
	}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("failed to start server")
		os.Exit(1)
	}
}
