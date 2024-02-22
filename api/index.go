package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"slices"
)

type pkg struct {
	Versions map[string]string
}

func Api(rw http.ResponseWriter, req *http.Request) {
	pkgs := map[string]pkg{}

	file, err := os.ReadFile("./all.json")
	if err != nil {
		http.Error(rw, "can't find cache file", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(file, &pkgs)
	if err != nil {
		http.Error(rw, "can't unmarshal cache file", http.StatusInternalServerError)
		return
	}

	pkgname := req.FormValue("package")
	version := req.FormValue("version")
	pkgVersions, ok := pkgs[pkgname]
	if !ok {
		http.Error(rw, "package not found", http.StatusNotFound)
		return
	}
	toReturn, ok := pkgVersions.Versions[version]
	if !ok {
		versions := []string{}
		for v := range pkgVersions.Versions {
			versions = append(versions, v)
		}
		slices.Sort(versions)
		http.Error(rw, fmt.Sprintf("version not found %v", versions), http.StatusNotFound)
		return
	}

	rw.Header().Set("Link", "https://github.com/NixOS/nixpkgs/archive/"+toReturn+".zip")
	rw.Header().Set("Location", "https://github.com/NixOS/nixpkgs/archive/"+toReturn+".zip")

	rw.WriteHeader(http.StatusMovedPermanently)
}