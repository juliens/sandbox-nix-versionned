package main

import (
	"log"
	"net/http"

	"github.com/juliens/sandbox-nix-versionned/pkg/handlers"
)

func main() {

	mux := &http.ServeMux{}
	mux.HandleFunc("/binary", handlers.Binary)
	mux.HandleFunc("/flake", handlers.Flake)
	mux.HandleFunc("/devshell", handlers.DevShell)
	log.Fatal(http.ListenAndServe(":8092", mux))
}
