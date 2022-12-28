{
    description = "ent-plus";
    inputs = {
      nixpkgs.url = "github:NixOS/nixpkgs/nixos-22.11";
      flake-utils.url = "github:numtide/flake-utils";
      gomod2nix.url = "github:nix-community/gomod2nix";
    };
    outputs = { self, nixpkgs, flake-utils, gomod2nix } :
      (flake-utils.lib.eachDefaultSystem
        (system:
          let
            pkgs = import nixpkgs { 
              inherit system;
              overlays = [ 
                gomod2nix.overlays.default
              ];
            };
            ent-server = pkgs.buildGoApplication {
              pname = "ent-server";
              version = "0.1.0";
              pwd = ./.;
              src = ./cmd/ent-server;
              modules = ./gomod2nix.toml;
            };
            ent-server-docker = pkgs.dockerTools.buildImage {
              name = "ent-server";
              config = { Cmd = [ "${ent-server}/bin/rust_nix_blog" ]; };
            };
          in {
            packages = {
              ent-server = ent-server;
              ent-server-docker = ent-server-docker;
            };
            defaultPackage = ent-server;
            devShell = 
              pkgs.mkShell {
                packages = [
                    pkgs.gomod2nix
                ];
                buildInputs = [
                    pkgs.go
                    pkgs.gopls
                ];
              };
          }));
}
