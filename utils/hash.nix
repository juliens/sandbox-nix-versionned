{ lib }:
let
  packages=builtins.fromJSON  (builtins.readFile ../all.json);
in
  name: version: let hash=packages.${name}.versions.${version}; in "github:nixos/nixpkgs/${hash}"
