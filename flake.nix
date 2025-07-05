{
  description = "A Go program to calibrate your sleep schedule";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "eepy";
          version = "0.1.0";
          src = ./.;
          vendorHash = "sha256-mknBOUz8Xpqz9Jm355uVbky5jmEsRjmk84WrCQ8txyk=";
        };

        devShells.default = pkgs.mkShell {
          buildInputs = [
            pkgs.go
          ];
        };
      });
}
