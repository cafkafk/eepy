{ pkgs, ... }:

pkgs.buildGoModule {
  pname = "eepy";
  version = "0.1.0";
  src = ../../.;
  subPackages = [ "cmd/eepy" ];
  vendorHash = null;
}
