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
            subPackages = [ "cmd/eepy" ];
            vendorHash = null;
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

              # Test with adjustment
              output = machine.succeed("eepy 10:00 --target 09:00 --adjustment 30m")
              assert "Day 1:" in output
              assert "Wake up at 10:00" in output
              assert "Day 2:" in output
              assert "Wake up at 09:30" in output
              assert "Day 3:" in output
              assert "Wake up at 09:00" in output

              # Test with complex adjustment
              output = machine.succeed("eepy 10:00 --target 05:00 --adjustment 3h45m")
              assert "Day 1:" in output
              assert "Wake up at 10:00" in output
              assert "Day 2:" in output
              assert "Wake up at 06:15" in output
              assert "Day 3:" in output
              assert "Wake up at 05:00" in output
            '';
          };
        };
    };
}
