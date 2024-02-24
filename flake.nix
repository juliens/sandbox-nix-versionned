
{
  description = "Test";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    go.url = "https://github.com/NixOS/nixpkgs/archive/e3ca01dedc1e.zip";
  };
  outputs = {nixpkgs,...}:
  let
    system = "x86_64-linux";
    pkgs = import nixpkgs {
      inherit system;
      config.allowUnfree = true;
    };
  in
  {
    packages.${system}.default = nixpkgs.legacyPackages.${system}.go_1_25;
    devShells.${system}.default = pkgs.mkShell {
      packages = [
        go.${system}.go_1_21
      ];
    };
  };
}
