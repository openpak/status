# Status Page

OpenPak's Status Page.

| | |
| --- | --- |
| Network | Switch (the console checks), plus the public site and account core they depend on |
| Role | Public status page: Gatus checks the network from a separate host, and `cmd/statuspage` renders the result |
| Depends on | Gatus (`ghcr.io/twin/gatus`), run as a container |
| Licence | AGPL-3.0-only |

## What it checks

Gatus runs three groups of checks: the public site, account core and console
link page (through Cloudflare); the console endpoints a console reaches through
its own DNS redirect; and the game session ports. The full list is in
[Protocol](docs/protocol.md), and the check set itself is in
[gatus.yaml](gatus.yaml).

## Build

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o statuspage ./cmd/statuspage
```

Ports and the deploy steps are in [Deploy](docs/deploy.md).

## Regenerating

The generated part of `gatus.yaml` and the public ports table both come from the
production quadlets (a copy of `~/.config/containers/systemd/` from the primary box):

```sh
go run ./cmd/gen-gatus -quadlets DIR -exclude openpak-nx-baas   # rewrites the block marked "generated" in gatus.yaml
go run ./cmd/gen-ports -quadlets DIR > ../../../../docs/docs/contributors/ports.md
```

`nx-baas` is excluded because its port 21080 is blocked by the cloud firewall (the Public group already covers it). A port bound to a specific address (loopback, ZeroTier) counts as internal and is not probed. Hand-written endpoints are never touched; re-running replaces only the generated block.

## Docs

- [Protocol](docs/protocol.md)
- [Deploy](docs/deploy.md)

PRDs: openpak/prds (private)

## Licence

AGPL-3.0-only. See [LICENSE](LICENSE).
