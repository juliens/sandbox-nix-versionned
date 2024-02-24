#!/usr/bin/env bash
if [ ! -s "/tmp/jsons/$1.json" ]; then
    if [ -f /tmp/jsons/$1.json ]; then
        rm /tmp/jsons/$1.json
    fi
else
    echo "Already exists $1"
    exit 0
fi

URL="git+file:///home/juliens/dev/nixpkgs?ref=$1"
HASH="nix flake prefetch $URL --json | jq .hash"


    echo "Try $1"
NIXPKGS_ALLOW_INSECURE=1 nix eval --impure --json $URL#legacyPackages.x86_64-linux --apply "
  builtins.mapAttrs  (
      name: value:
      let
        version = (builtins.tryEval value.version or \"\");
        mainProgram = (builtins.tryEval value.meta.mainProgram or name);
      in
      if (builtins.isString version.value)
      then
      (
        {
          versions.\${version.value}={ commit=\"$1\"; lock=\"$HASH\"; };
          mainProgram=mainProgram.value;
        }
      )
      else
      (
        {}
      )
    )
" > /tmp/jsons/$1.json

