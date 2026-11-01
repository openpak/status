# Protocol: status

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
| Xbox 360 | `/xmatch/xmatchclient2.ashx` on the origin | matchmaking (Live service 6) |
| Xbox 360 | `/xstats/xstatread.ashx` on the origin | leaderboards (Live service 7) |
| Xbox 360 | `/XArb/Register.ashx` on the origin | ranked-match arbitration (Live service 16) |

The public checks go through Cloudflare on purpose, so a pass means what a
person experiences works — not merely that a process is alive.

The console names resolve to **Nintendo** for everyone else. A console only
reaches us because its own DNS is redirected, so the container maps those four
names to the origin IP (`PodmanArgs=--add-host`) and the checks follow the
console's path. TLS verification is off for them because those certificates are
signed by the OpenPak CA, which is the entire point of them.

The Xbox 360 checks POST to the origin's Traefik without the gateway key, so the
service refuses with 403 before reading anything: 403 means routed and running,
404 a missing route, 502 a stopped container.
