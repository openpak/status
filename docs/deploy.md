# Deploy: status

## Deploying

`<secondary-host>` is the status box's SSH login (see the private deploy notes).

```sh
scp gatus.yaml <secondary-host>:~/openpak-status/gatus.yaml
scp openpak-status.container <secondary-host>:~/.config/containers/systemd/
ssh <secondary-host> 'systemctl --user daemon-reload && systemctl --user restart openpak-status'
```

Two things about that host, both discovered the hard way:

- **Docker Hub is unreachable from it** — `registry-1.docker.io` times out,
  while ghcr and quay answer normally. The image comes from
  `ghcr.io/twin/gatus`.
- **podman is 4.9**, where the `AddHost=` quadlet key does not exist yet. Hence
  `PodmanArgs=--add-host=…`. An `apt full-upgrade` will not change this;
  Ubuntu 24.04 ships 4.9.

## Two processes, and which one is public

| Port | Process | Reachable from |
| --- | --- | --- |
| 8450 | `statuspage` — the OpenPak-branded page and `/api/status` | the internet, via the tunnel |
| 8460 | Gatus — the checker, its scheduling, retries and history | localhost only |

**Live at `status.openpak.org`.** The tunnel on that host is token-managed, so
its routes live in the Cloudflare dashboard rather than on disk; the public
hostname points at `http://localhost:8450`. If the page ever shows Gatus's own
dashboard instead of this one, the tunnel is pointed at the checker.

## Building the page

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o statuspage ./cmd/statuspage
scp statuspage <secondary-host>:~/openpak-status/statuspage
scp openpak-statuspage.service <secondary-host>:~/.config/systemd/user/
ssh <secondary-host> 'systemctl --user daemon-reload && systemctl --user restart openpak-statuspage'
```

A plain binary and a systemd user unit rather than a container: the deploy host
cannot reach Docker Hub, and this needs nothing an image would provide.
