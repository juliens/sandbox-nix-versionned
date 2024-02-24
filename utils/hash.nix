{ lib }:
let
  Packages=builtins.fromJSON  (builtins.readFile ../all.json);
in
  name: version: let hash=Packages.${name}.versions.${version}; in "github:nixos/nixpkgs/${hash}"
