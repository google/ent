{ pkgs ? import <nixpkgs> {} }:
  pkgs.mkShell {
    nativeBuildInputs = [
        pkgs.nodePackages.prettier
        pkgs.buildPackages.go
    ];
}
