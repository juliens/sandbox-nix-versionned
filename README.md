# sandbox-nix-versionned

## devShell


```
{
    "name":"my devshell",
    "nixpkgs":"github:nixos/nixpkgs",
    "packages": {
        "jq":"*",
        "go":"1.20.*",
        ...
    }
}
```

```
curl -d@env.json https://nix.juguul.ovh/api/devshell -o flake.nix
```

