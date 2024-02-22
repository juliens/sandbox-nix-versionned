package handler

import (
	"net/http"

	"github.com/juliens/sandbox-nix-versionned/pkg"
)

func Api(rw http.ResponseWriter, req *http.Request) {
	url, err := pkg.GetURL(req.FormValue("package"), req.FormValue("version"))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Link", url)
	rw.Header().Set("Location", url)

	rw.WriteHeader(http.StatusMovedPermanently)
}
