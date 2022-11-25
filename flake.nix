{
    description = "ent-plus";
    inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-22.05";
    outputs = { self, nixpkgs } : let
        pkgs = nixpkgs.legacyPackages.x86_64-linux;
        in {
            devShell.x86_64-linux = 
              pkgs.mkShell {
                buildInputs = [
                    pkgs.go
                    pkgs.gopls
                ];
              };
        };
}
