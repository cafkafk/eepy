# SPDX-FileCopyrightText: 2025 Christina Sørensen
#
# SPDX-License-Identifier: EUPL-1.2

{
  pkgs,
  eepy,
  ...
}:

pkgs.nixosTest {
  name = "eepy-test";
  nodes.machine = {
    environment.systemPackages = [ eepy ];
  };
  testScript = ''
    machine.wait_for_unit("multi-user.target")
    machine.succeed("rm -f /root/.config/eepy/plan.json")
    output = machine.succeed("eepy 08:00")
    assert "Your sleep calibration plan:" in output
    assert "Wake up at 08:00" in output
    assert "Go to bed at 23:00" in output

    # Test with adjustment
    machine.succeed("rm -f /root/.config/eepy/plan.json")
    output = machine.succeed("eepy 10:00 --target 09:00 --adjustment 30m")
    assert "(Day 1):" in output
    assert "Wake up at 10:00" in output
    assert "(Day 2):" in output
    assert "Wake up at 09:30" in output
    assert "(Day 3):" in output
    assert "Wake up at 09:00" in output

    # Test with complex adjustment
    machine.succeed("rm -f /root/.config/eepy/plan.json")
    output = machine.succeed("eepy 10:00 --target 05:00 --adjustment 3h45m")
    assert "(Day 1):" in output
    assert "Wake up at 10:00" in output
    assert "(Day 2):" in output
    assert "Wake up at 06:15" in output
    assert "(Day 3):" in output
    assert "Wake up at 05:00" in output

    # Test with start date
    machine.succeed("rm -f /root/.config/eepy/plan.json")
    output = machine.succeed("eepy 08:00 --start-date 2025-01-01")
    assert "Your sleep calibration plan:" in output
    assert "Wed, Jan 1 (Day 1):" in output
    assert "Wake up at 08:00" in output
'';
}
