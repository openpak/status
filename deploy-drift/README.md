# deploy-drift

`openpak-drift-check` answers one question daily on each box: is every service that has a release running it?
It never rolls anything.

| File | Goes to |
| --- | --- |
| `openpak-drift-check` | `~/.local/bin/openpak-drift-check` (main and secondary) |
| `openpak-drift-check.service` / `.timer` | `~/.config/systemd/user/` (daily 07:15 UTC; `OnFailure=openpak-unit-failed@%n`; the secondary sets `DRIFT_ROLE=secondary`) |
| `drift-allowlist.txt` | main `~/openpak/deploy/`: container packages deliberately not deployed (libraries, clients, emulators), `name  # reason` |
| `secondary-units.txt` | main `~/openpak/deploy/`: image packages that run on the secondary box (keep in step with its quadlets) |

Checks: (1) each `ghcr.io/openpak/*` quadlet's running digest against the newest `v*.*.*` release image on ghcr;
a `# pinned: <reason>` comment in the quadlet lists the unit as pinned and skips it; a fixed tag or digest without that comment is flagged immediately, even while current (everything follows `:latest`). (2) main only: every org container
package must back a quadlet, be in `secondary-units.txt` or be in the allowlist. Oneshot and manual units are compared
by the image on disk. Each flagged unit says why `podman-auto-update` did not catch it (digest pin, fixed tag, no
`AutoUpdate=registry`, `:latest` not moved by the release, or the auto-update timer has not run / is disabled).

Report: `~/openpak/deploy/drift-report.txt`. Exit 1 when anything is flagged, so the unit fails and the OnFailure alert
fires. By hand: `openpak-drift-check` (takes about a minute).
