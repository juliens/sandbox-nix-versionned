{
  description="test";
  inputs = 
  #packages=builtins.fromJSON  (builtins.readFile ./all.json);
  #name="go";
  #version="1.20.3";
  #gohash=packages.${name}.versions.${version};
  {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    #gopkgs.url =  "github:nixos/nixpkgs/${gohash}";
    gopkgs.url =  "github:nixos/nixpkgs/80f198ff3a65";
  };
  outputs = {self, nixpkgs, gopkgs}: {
    packages.x86_64-linux.default=gopkgs.legacyPackages.x86_64-linux.go;
  };
}
