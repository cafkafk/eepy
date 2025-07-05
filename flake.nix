{
  description = "A Go program to calibrate your sleep schedule";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs = { self, nixpkgs, flake-parts, ... }@inputs:
    flake-parts.lib.mkFlake { inherit self inputs; } {
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      perSystem = { pkgs, ... }:
        let
          eepy-pkg = pkgs.buildGoModule {
            pname = "eepy";
            version = "0.1.0";
            src = ./.;
            vendorHash = "sha256-mknBOUz8Xpqz9Jm355uVbky5jmEsRjmk84WrCQ8txyk=";
          };
        in
        {
          packages.default = eepy-pkg;

          devShells.default = pkgs.mkShell {
            buildInputs = [
              pkgs.go
            ];
          };

          checks.default = pkgs.nixosTest {
            name = "eepy-test";
            nodes.machine = {
              environment.systemPackages = [ eepy-pkg ];
            };
            testScript = ''
              machine.wait_for_unit("multi-user.target")
              output = machine.succeed("eepy 08:00")
              assert "Your sleep calibration plan:" in output
              assert "Wake up at 08:00" in output
              assert "Go to bed at 23:00" in output
            '';
          };
        };
    };
}
