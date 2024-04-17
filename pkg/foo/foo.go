package foo

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/Masterminds/semver"
	"github.com/juliens/sandbox-nix-versionned/pkg/handlers"
	"github.com/juliens/sandbox-nix-versionned/pkg/nix"
)

//go:embed all.json
var b []byte

type fileStruct struct {
	Packages map[string]nix.Pkg
	Commit   map[string]string
}

type Foo struct {
	path       string
	output     fileStruct
	outputLock sync.RWMutex
}

func NewInternal() (*Foo, error) {
	output := fileStruct{
		Commit:   map[string]string{},
		Packages: map[string]nix.Pkg{},
	}

	err := json.Unmarshal(b, &output)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshaling previous path: %w", err)

	}

	return &Foo{
		output: output,
	}, nil
}

func New(filepath string) (*Foo, error) {
	output := fileStruct{
		Commit:   map[string]string{},
		Packages: map[string]nix.Pkg{},
	}

	file, err := os.ReadFile(filepath)
	if err == nil {
		err := json.Unmarshal(file, &output)
		if err != nil {
			return nil, fmt.Errorf("error while unmarshaling previous path %s: %w", filepath, err)
		}
	}

	return &Foo{
		path:   filepath,
		output: output,
	}, nil
}

func (f *Foo) Merge(packages map[string]nix.Pkg) {
	f.outputLock.Lock()
	defer f.outputLock.Unlock()

	for pkgName, pkgData := range packages {
		if previousPkgData, ok := f.output.Packages[pkgName]; ok {
			if previousPkgData.Versions == nil {
				previousPkgData.Versions = map[string]nix.Version{}
			}
			for v, d := range pkgData.Versions {
				previousPkgData.Versions[v] = d
			}
			f.output.Packages[pkgName] = previousPkgData
		} else {
			f.output.Packages[pkgName] = pkgData
		}
	}
}

func (f *Foo) AddCommit(commit string, err error) {
	f.outputLock.Lock()
	defer f.outputLock.Unlock()

	if err == nil {
		f.output.Commit[commit] = "Handled"
		return
	}
	f.output.Commit[commit] = err.Error()
}

func (f *Foo) ContainsCommit(commit string) bool {
	f.outputLock.RLock()
	defer f.outputLock.RUnlock()

	_, ok := f.output.Commit[commit]
	return ok
}

func (f *Foo) Write() error {
	if len(f.path) == 0 {
		return errors.New("internal version")
	}

	f.outputLock.RLock()
	defer f.outputLock.RUnlock()

	marshal, err := json.Marshal(f.output)
	if err != nil {
		return err
	}

	return os.WriteFile(f.path, marshal, 0777)
}

func (f *Foo) GetPackageVersionnedFlakeURL(pkgName, version string) (string, error) {
	ver, err := f.GetPackageVersionned(pkgName, version)
	if err != nil {
		return "", err
	}

	return f.getFlakeUrl(ver), nil
}

func (f *Foo) getFlakeUrl(version nix.Version) string {
	return "https://github.com/NixOS/nixpkgs/archive/" + version.Commit + ".tar.gz"
}

func (f *Foo) GetBinaryVersionnedFlakeURL(binaryName, version string) (string, error) {
	_, ver, err := f.GetBinaryVersionned(binaryName, version)
	if err != nil {
		return "", err
	}

	return f.getFlakeUrl(ver), nil

}

func (f *Foo) GetPackageVersionned(pkgName, version string) (nix.Version, error) {
	f.outputLock.RLock()
	defer f.outputLock.RUnlock()

	pkgVersions, ok := f.output.Packages[pkgName]
	if !ok {
		return nix.Version{}, errors.New("package not found")
	}

	toReturn, ok := pkgVersions.Versions[version]
	if !ok {
		versions := semver.Collection{}
		oldVersions := map[*semver.Version]nix.Version{}
		for v, value := range pkgVersions.Versions {
			newVersion, err := semver.NewVersion(v)
			if err != nil {
				continue
			}
			value.Version = v
			oldVersions[newVersion] = value
			versions = append(versions, newVersion)
		}

		sort.Sort(versions)
		slices.Reverse(versions)
		c, err := semver.NewConstraint(version)
		if err != nil {
			return nix.Version{}, err
		}

		if pkgName == "kubectl" {
			fmt.Println(versions)
		}
		for _, v := range versions {
			if c.Check(v) {
				fmt.Println("Selected version", pkgName, v.String())
				return oldVersions[v], nil
			}
		}

		return nix.Version{}, fmt.Errorf("version not found %v", versions)
	}

	toReturn.Version = version
	return toReturn, nil
}

func (f *Foo) GetBinaryVersionned(binaryName, ver string) (string, nix.Version, error) {
	f.outputLock.RLock()
	defer f.outputLock.RUnlock()

	if version, err := f.GetPackageVersionned(binaryName, ver); err == nil {
		return binaryName, version, nil
	}

	for pkgName, pkg := range f.output.Packages {
		if pkg.MainProgram != binaryName {
			continue
		}
		if version, err := f.GetPackageVersionned(pkgName, ver); err == nil {
			return pkgName, version, nil
		}

	}
	return "", nix.Version{}, fmt.Errorf("%s version %s not found", binaryName, ver)
}

func (f *Foo) GetDevShellFlakeFile(config handlers.DevShellConfig) ([]byte, error) {
	if config.Nixpkgs == "" {
		config.Nixpkgs = "github.com:nixos/nixpkgs"
	}

	if config.Name == "" {
		config.Name = "devshell"
	}

	var inputs []string
	var pkgs []string
	for binaryName, version := range config.Packages {
		pkgName, ver, err := f.GetBinaryVersionned(binaryName, version)
		if err != nil {
			return nil, err
		}
		if version == "*" {
			pkgs = append(pkgs, pkgName)
			continue
		}

		inputs = append(inputs, "    "+binaryName+".url=\""+f.getFlakeUrl(ver)+"\"; # "+pkgName+" - "+ver.Version)
		pkgs = append(pkgs, "    inputs."+binaryName+".legacyPackages.${system}."+pkgName)
	}
	template := `{
  description = "` + config.Name + `";
  inputs = {
    nixpkgs.url = "` + config.Nixpkgs + `";
    flake-utils.url = "github:numtide/flake-utils";

    ` + strings.Join(inputs, "\n") + `
  };
  outputs = inputs @ {nixpkgs, flake-utils,...}:
  flake-utils.lib.eachDefaultSystem (system: let
    pkgs = import nixpkgs {
      inherit system;
      config.allowUnfree = true;
    };
  in
  {
    devShells.default = pkgs.mkShell {
      packages = with pkgs; [
        ` + strings.Join(pkgs, "\n        ") + `		
      ];
    };
  });
}
`
	return []byte(template), nil
}

func (f *Foo) GetBinaryFlakeTarGz(rw io.Writer, binaryName, version string) ([]byte, error) {
	pkgName, ver, err := f.GetBinaryVersionned(binaryName, version)
	if err != nil {
		return nil, err
	}
	flakeTemplate := `
{
	description = "Test";
	inputs = {
		flake-utils.url = "github:numtide/flake-utils";
		nixpkgs.url = "` + f.getFlakeUrl(ver) + `";
	};
	outputs = {nixpkgs,...}:
	flake-utils.lib.eachDefaultSystem (system: let
	  pkgs = import nixpkgs {
        inherit system;
       config.allowUnfree = true;
      };
	in
	{
		packages.default = pkgs.` + pkgName + `;
	});
}
`

	lockTemplate := `
{
  "nodes": {
    "nixpkgs": {
      "locked": {
        "narHash": "` + ver.Lock + `",
        "type": "tarball",
        "url": "` + f.getFlakeUrl(ver) + `"
      },
      "original": {
        "type": "tarball",
        "url": "` + f.getFlakeUrl(ver) + `"
      }
    },
    "root": {
      "inputs": {
        "nixpkgs": "nixpkgs"
      }
    }
  },
  "root": "root",
  "version": 7
}`

	_ = lockTemplate

	return writeTarGz(rw, map[string][]byte{
		"flake.nix":  []byte(flakeTemplate),
		"flake.lock": []byte(lockTemplate),
	})
}

func writeTarGz(rw io.Writer, files map[string][]byte) ([]byte, error) {
	temp, err := os.MkdirTemp("/tmp", "flake")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(temp)

	gw := gzip.NewWriter(rw)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	for name, content := range files {
		err = os.WriteFile("./"+name, content, 0777)
		if err != nil {
			return nil, err
		}

		stat, err := os.Stat("./" + name)
		if err != nil {
			return nil, err
		}

		header, err := tar.FileInfoHeader(stat, "./"+name)
		if err != nil {
			return nil, err
		}

		header.Name = "/nixpkgs/" + name
		err = tw.WriteHeader(header)
		if err != nil {
			return nil, err
		}

		// Copy file content to tar archive
		r := bytes.NewReader(content)
		_, err = io.Copy(tw, r)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}
