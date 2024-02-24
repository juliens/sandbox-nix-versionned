package handler

import (
	"net/http"

	"github.com/juliens/sandbox-nix-versionned/pkg/handlers"
)

func DevShell(rw http.ResponseWriter, req *http.Request) {
	handlers.DevShell(rw, req)
}
