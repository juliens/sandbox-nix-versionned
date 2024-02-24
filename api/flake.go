package handler

import (
	"net/http"

	"github.com/juliens/sandbox-nix-versionned/pkg/handlers"
)

func Flake(rw http.ResponseWriter, req *http.Request) {
	handlers.Flake(rw, req)
}
