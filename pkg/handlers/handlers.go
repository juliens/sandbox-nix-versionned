package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/juliens/sandbox-nix-versionned/pkg/foo"
)

type DevShellConfig struct {
	Name     string            `json:"name"`
	Nixpkgs  string            `json:"nixpkgs"`
	Packages map[string]string `json:"packages"`
}

func DevShell(rw http.ResponseWriter, req *http.Request) {
	n, err := foo.NewInternal()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	decoder := json.NewDecoder(req.Body)
	binaries := DevShellConfig{}
	err = decoder.Decode(&binaries)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	flake, err := n.GetDevShellFlakeFile(binaries)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}
	rw.Write(flake)

}

func Binary(rw http.ResponseWriter, req *http.Request) {
	binaryName := req.FormValue("binary")
	version := req.FormValue("version")
	n, err := foo.NewInternal()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	_, err = n.GetBinaryFlakeTarGz(rw, binaryName, version)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

}

func Flake(rw http.ResponseWriter, req *http.Request) {
	f, err := foo.NewInternal()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}
	pkgName := req.FormValue("package")
	version := req.FormValue("version")

	url, err := f.GetPackageVersionnedFlakeURL(pkgName, version)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	rw.Header().Set("Link", url)
	rw.Header().Set("Location", url)

	rw.WriteHeader(http.StatusMovedPermanently)
}
