package handler

import (
	"net/http"

	"github.com/juliens/sandbox-nix-versionned/pkg/handlers"
)

func Binary(rw http.ResponseWriter, req *http.Request) {
	handlers.Binary(rw, req)
}
