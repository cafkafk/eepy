_:

{
  perSystem =
    { pkgs, ... }:
    {
      pre-commit.settings = {
        hooks = {
          # Nix
          nixfmt-rfc-style.enable = true;

          # Flake check
          flake-checker.enable = true;

          # Incremental analysis assistant for writing in Nix.
          nil.enable = true;

          reuse = {
            # TODO: Make reuse compliant
            enable = false;
            name = "reuse";
            entry = with pkgs; "${reuse}/bin/reuse lint";
            pass_filenames = false;
          };

          gitleaks = {
            enable = true;
            name = "gitleaks";
            description = "Scan for hardcoded secrets";
            entry = "gitleaks detect --source . --verbose"; # The command to run
            files = ".*";
            language = "system";
            stages = [ "pre-commit" ];
            # Make gitleaks available to the hook's environment
            package = pkgs.gitleaks;
          };
        };
      };
    };
}
