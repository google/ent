{
    description = "ent-plus";
    inputs = {
      nixpkgs.url = "github:NixOS/nixpkgs/nixos-22.05";
      flake-utils.url = "github:numtide/flake-utils";
      gomod2nix.url = "github:nix-community/gomod2nix";
    };
    outputs = { self, nixpkgs, flake-utils, gomod2nix } : flake-utils.lib.eachDefaultSystem(system:
      let pkgs = import nixpkgs { 
        inherit system;
        ovelays = [ gomod2nix.overlays.default ];
      };
      in rec {
        packages = {
          ent-server = pkgs.buildGoModule {
            name = "ent-plus";
            src = ./.;
            vendorSha256 = "sha256-pQpattmS9VmO3ZIQUFn66az8GSmB4IvYhTTCFn6SUmo=";
          };
        };
        defaultPackage = packages.ent-server;
        devShell = 
          pkgs.mkShell {
            packages = [
                # pkgs.gomod2nix
            ];
            buildInputs = [
                pkgs.go
                pkgs.gopls
            ];
          };
      });
}
