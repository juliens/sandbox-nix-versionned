
{
	description = "Test";
	inputs = {
		nixpkgs.url = "https://github.com/NixOS/nixpkgs/archive/dfccc488dbbc.zip";
	};
	outputs = {nixpkgs,...}:{
		packages.x86_64-linux.default = nixpkgs.legacyPackages.x86_64-linux.go;
	};
}
