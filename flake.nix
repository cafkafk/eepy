{
  description = "A Go program to calibrate your sleep schedule";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    gitignore = {
      url = "github:hercules-ci/gitignore.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    flake-compat = {
      url = "github:inclyc/flake-compat";
      flake = false;
    };

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    git-hooks-nix = {
      url = "github:cachix/git-hooks.nix";
      inputs = {
        nixpkgs.follows = "nixpkgs";
        flake-compat.follows = "flake-compat";
        gitignore.follows = "gitignore";
      };
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-parts,
      ...
    }@inputs:
    flake-parts.lib.mkFlake { inherit self inputs; } {
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      imports = [
        inputs.git-hooks-nix.flakeModule
      ];

      perSystem =
        { pkgs, config, ... }:
        let
          eepy = pkgs.callPackage ./nix/package/eepy.nix { inherit pkgs; };
        in
        {
          packages.default = eepy;

          devShells.default = pkgs.mkShell {
            buildInputs = [
              pkgs.go
            ];
            inputsFrom = [ config.pre-commit.devShell ];
          };

          checks.default = pkgs.callPackage ./nix/vm_tests/integration_tests.nix { inherit pkgs eepy; };
        };
    };
}
