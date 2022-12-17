{
    description = "ent-plus";
    inputs = {
      nixpkgs.url = "github:NixOS/nixpkgs/nixos-22.05";
      flake-utils.url = "github:numtide/flake-utils";
      gomod2nix.url = "github:nix-community/gomod2nix";
    };
    outputs = { self, nixpkgs, flake-utils, gomod2nix } : (flake-utils.lib.eachDefaultSystem(system:
      let
        pkgs = import nixpkgs { 
          inherit system;
          ovelays = [ 
            gomod2nix.overlays.default
          ];
        };
        # ent-server = pkgs.buildGoApplication {
        #   name = "ent-plus";
        #   src = ./.;
        # };
        ent-server = pkgs.hello;
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
                # pkgs.gomod2nix
            ];
            buildInputs = [
                # gomod2nix.packages.${system}.default
                pkgs.go
                pkgs.gopls
            ];
          };
      }));
}
