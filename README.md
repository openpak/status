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

## Making it public

The page listens on `127.0.0.1:8450`; the Cloudflare tunnel already on that
host publishes it. The tunnel is token-managed, so its routes live in the
Cloudflare dashboard rather than on disk: add a public hostname
`status.openpak.org` pointing at `http://localhost:8450` under Zero Trust →
Networks → Tunnels. Cloudflare creates the DNS record itself.
