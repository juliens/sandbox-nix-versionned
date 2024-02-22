package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
)

type pkg struct {
	Versions map[string]string
}

func main() {
	pkgs := map[string]pkg{}

	file, err := os.ReadFile("./all.json")
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(file, &pkgs)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe(":8091", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
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
	})))

}
