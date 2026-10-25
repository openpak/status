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

## Docs

- [Protocol](docs/protocol.md)
- [Evidence](docs/evidence.md)
- [Deploy](docs/deploy.md)
- [Notes](docs/notes.md)
- [Docs index](docs/README.md)

## Licence

AGPL-3.0-only. See [LICENSE](LICENSE).
