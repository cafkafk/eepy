# SPDX-FileCopyrightText: 2025 Christina Sørensen
#
# SPDX-License-Identifier: EUPL-1.2

{ pkgs, ... }:

pkgs.buildGoModule {
  pname = "eepy";
  version = "0.1.0";
  src = ../../.;
  subPackages = [ "cmd/eepy" ];
  vendorHash = null;
}
