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
| Xbox 360 | `/xuacs/PassportMemberNameExists.ashx` on the origin | accounts and profile recovery (Live service 5) |
| Xbox 360 | `/xpnfront/xpresence.srf` on the origin | presence and friends (Live service 1) |
| Xbox 360 | `/msgserver/vetstring2.ashx` on the origin | profile text check (Live service 2) |
| Xbox 360 | `/xedl/ProfileInfo/103/CountryInfo.cib` on the origin | country information (Live service 11) |
| Xbox 360 | `/xsig/ValidateCertificate.ashx` on the origin | certificate validation (Live service 12); needs the WLID CA from xuacs |
| Xbox 360 | `/xbos/GetRevocationList.ashx` on the origin | revocation list (Live service 21) |
| Xbox 360 | `/xstats/readsettings.ashx` on the origin | profile settings (Live service 34) |
| Xbox 360 | `/xmatch/xmatchclient2.ashx` on the origin | matchmaking (Live service 6) |
| Xbox 360 | `/xstats/xstatread.ashx` on the origin | leaderboards (Live service 7) |
| Xbox 360 | `/XArb/Register.ashx` on the origin | ranked-match arbitration (Live service 16) |
| PS3 | `ranking-view.np.community.playstation.net/ranking_view/func/get_ranking` | a leaderboard read (np-score and its database) |
| PS3 | `static-resource…/np/resource/title/<commid>/storage/…-storage.xml` | the signed Title User Storage discovery (np-tus) |
| PS3 | `nsx.sec.np.dl.playstation.net/nsx/…/NSXWSV-PN.P3.US-WHATSNEW01.xml` | the XMB What's New feed (np-community) |
| PS3 | `auth.np.ac.playstation.net/2.0/oauth/token` | sign-in: np-account's OAuth handler answering an empty request |
| PS3 | `getprof.us.np.community.playstation.net/basic_view/func/get_profile` | profiles: np-community's handler rejecting a malformed JID |
| PS3 | `145.241.199.19:5223` (TCP) | friends and presence: the np-community XMPP stream accepts connections |
| Original Xbox | `145.241.228.207:3074` (UDP), `88/udp` on the primary | gateway and Kerberos KDC accept datagrams |
| Original Xbox | `macs.xboxlive.com` A query to xb-dns on the primary ZeroTier address | discovery DNS answers with the auth host |
| Original Xbox | `10.212.63.196:26111/26112/26115/26117` (TCP) | accounts, terms, friends and presence, voice: mutual-TLS APIs accept connections from the status host |

The public checks go through Cloudflare on purpose, so a pass means what a
person experiences works — not merely that a process is alive.

The console names resolve to **Nintendo** for everyone else. A console only
reaches us because its own DNS is redirected, so the container maps those four
names to the origin IP (`PodmanArgs=--add-host`) and the checks follow the
console's path. TLS verification is off for them because those certificates are
signed by the OpenPak CA, which is the entire point of them.

The Xbox 360 checks reach the origin's Traefik without the gateway key, so the
identity headers are dropped: services that need them refuse with 403, the rest
answer the empty request their own way (400, or the real reply). Any of those
means routed and running; 404 is a missing route, 502 a stopped container.
