{
  description = "A Go program to calibrate your sleep schedule";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs = { self, nixpkgs, flake-parts, ... }@inputs:
    flake-parts.lib.mkFlake { inherit self inputs; } {
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      perSystem = { pkgs, ... }: {
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
      };
    };
}
