package pkg

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
)

//go:embed all.json
var b []byte

type pkg struct {
	Versions map[string]string
}

func GetURL(pkgName, version string) (string, error) {
	pkgs := map[string]pkg{}

	err := json.Unmarshal(b, &pkgs)
	if err != nil {
		return "", errors.New("can't unmarshal cache file")
	}

	pkgVersions, ok := pkgs[pkgName]
	if !ok {
		return "", errors.New("package not found")
	}

	toReturn, ok := pkgVersions.Versions[version]
	if !ok {
		versions := []string{}
		for v := range pkgVersions.Versions {
			versions = append(versions, v)
		}
		slices.Sort(versions)
		return "", fmt.Errorf("version not found %v", versions)
	}
	return "https://github.com/NixOS/nixpkgs/archive/" + toReturn + ".zip", nil
}
