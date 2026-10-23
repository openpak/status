# OpenPak status

The network's status page. It runs on **Draughts**, deliberately not on the
production box: a status page that lives on the machine it watches reports "up"
right until the moment it cannot report at all, which is the only moment anyone
reads it.

## What it checks

| Group | Endpoint | Why |
|---|---|---|
| Public | `openpak.org` | the site, through Cloudflare, as a visitor reaches it |
| Public | `account.openpak.org/healthz` | the account core |
| Public | `link.openpak.org/link` | the page a phone opens to link a console |
| Console | `dauth-lp1.../v8/challenge` | device auth; nothing a console does works without it |
| Console | `…baas.nintendo.com/1.0.0/certificates` | the key set every console token is verified against |
| Console | `accounts.nintendo.com/connect/1.0.0/authorize` | the sign-in page a console shows |
| Console | `dragons.../v2/elicenses/exercise` | the licence check; a 404 here stops account setup |

The public checks go through Cloudflare on purpose, so a pass means what a
person experiences works — not merely that a process is alive.

The console names resolve to **Nintendo** for everyone else. A console only
reaches us because its own DNS is redirected, so the container maps those four
names to the origin IP (`PodmanArgs=--add-host`) and the checks follow the
console's path. TLS verification is off for them because those certificates are
signed by the OpenPak CA, which is the entire point of them.

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

Gatus does the monitoring and stays private; `cmd/statuspage` renders what it
found in the site's own colours and words. None of the hard parts of monitoring
are reimplemented, and none of the presentation is left to a tool that does not
know what any of these services are for.

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
