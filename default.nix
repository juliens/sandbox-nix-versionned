{
  nixpkgs ? import <nixpkgs> {}
}:
let
  hashgo=(nixpkgs.callPackage ./utils/hash.nix { lib=nixpkgs.lib; }) "go" "1.20.3";
in
  hashgo
