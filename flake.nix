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
        ent-server = pkgs.buildGoApplication {
          name = "ent-plus";
          src = ./.;
        };
        # ent-server = pkgs.hello;
        # dockerImage = pkgs.dockerTools.buildImage {
        #   name = "rust-nix-blog";
        #   config = { Cmd = [ "${ent-server}/bin/rust_nix_blog" ]; };
        # };
      in {
        packages = {
          ent-server = ent-server;
          # dockerImage = dockerImage;
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