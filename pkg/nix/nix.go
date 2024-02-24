package nix

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type Nix struct {
	Url string
}

func NewNix(path string) *Nix {
	return &Nix{Url: "git+file://" + path}

}

type PrefetchData struct {
	Hash      string `json:"hash"`
	StorePath string `json:"storePath"`
}

func (n *Nix) prefetch(ctx context.Context, commit string) (*PrefetchData, error) {
	url := n.Url + "?ref=" + commit
	prefetchCmd := exec.CommandContext(ctx, "nix", "flake", "prefetch", url, "--json")

	output, err := prefetchCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error while prefetching data %q: %w", output, err)
	}

	prefetchData := &PrefetchData{}
	err = json.Unmarshal(output, prefetchData)
	if err != nil {
		return nil, err
	}
	return prefetchData, nil
}

type Pkg struct {
	Versions    map[string]Version `json:"versions"`
	MainProgram string             `json:"mainProgram"`
}

type Version struct {
	Commit  string `json:"Commit"`
	Lock    string `json:"lock"`
	Version string `json:"-"`
}

func (n *Nix) Packages(ctx context.Context, commit string) (map[string]Pkg, error) {
	prefetchData, err := n.prefetch(ctx, commit)
	if err != nil {
		return nil, err
	}

	evalData, err := n.eval(ctx, commit, prefetchData.Hash)
	if err != nil {
		return nil, err
	}

	exec.CommandContext(ctx, "nix", "store", "delete", prefetchData.StorePath).Start()

	return evalData, nil
}

func (n *Nix) eval(ctx context.Context, commit, hash string) (map[string]Pkg, error) {
	program := `
  builtins.mapAttrs  (
      name: value:
      let
        version = (builtins.tryEval value.version or "");
        mainProgram = (builtins.tryEval value.meta.mainProgram or name);
      in
      if (builtins.isString version.value)
      then
      (
        {
          versions.${version.value}={ Commit="` + commit + `"; lock="` + hash + `"; };
          mainProgram=mainProgram.value;
        }
      )
      else
      (
        {}
      )
    )
`

	url := n.Url + "?ref=" + commit
	evalCmd := exec.CommandContext(ctx, "nix", "eval", "--impure", "--json", url+"#legacyPackages.x86_64-linux", "--apply", program)
	evalCmd.Env = append(os.Environ(), "NIXPGS_ALLOW_INSECURE=1")

	output, err := evalCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error while eval packages %q: %w", output, err)
	}

	evalData := map[string]Pkg{}
	err = json.Unmarshal(output, &evalData)
	if err != nil {
		return nil, err
	}

	return evalData, nil
}
