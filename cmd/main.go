package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/juliens/sandbox-nix-versionned/pkg"
)

func main() {
	log.Println("????")
	err := http.ListenAndServe(":8092", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		url, err := pkg.GetURL(req.FormValue("package"), req.FormValue("version"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		flakeTemplate := `
{
	description = "Test";
	inputs = {
		nixpkgs.url = "` + url + `";
	};
	outputs = {nixpkgs,...}:{
		packages.x86_64-linux.default = nixpkgs.legacyPackages.x86_64-linux.go;
	};
}
`
		err = os.WriteFile("./flake.nix", []byte(flakeTemplate), 0777)
		if err != nil {
			http.Error(rw, "write file", http.StatusInternalServerError)
			return
		}
		gw := gzip.NewWriter(rw)
		defer gw.Close()
		tw := tar.NewWriter(gw)
		defer tw.Close()

		stat, err := os.Stat("./flake.nix")
		if err != nil {
			http.Error(rw, "stat file", http.StatusInternalServerError)
			return
		}

		header, err := tar.FileInfoHeader(stat, "./flake.nix")
		if err != nil {
			http.Error(rw, "header file", http.StatusInternalServerError)
			return
		}

		header.Name = "/nixpkgs/flake.nix"
		err = tw.WriteHeader(header)
		if err != nil {
			http.Error(rw, "write header", http.StatusInternalServerError)
			return
		}

		// Copy file content to tar archive
		r := bytes.NewReader([]byte(flakeTemplate))
		_, err = io.Copy(tw, r)
		if err != nil {
			http.Error(rw, "copy file", http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusOK)
	}))
	if err != nil {
		log.Println("boom")
	}
	log.Println("bizarre")
}
