# Protocol: status

## What it checks

| Group | Endpoint | Why |
|---|---|---|
| Public | `openpak.org` | the site, through Cloudflare, as a visitor reaches it |
| Core | `api.openpak.org/healthz/core` | the account core |
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
| PS3 | `a0.ww.np.dl.playstation.net/tpl/np/<title>/<title>-ver.xml` | game update checks: np-update's empty "no update" 404 |
| PS3 | `uploader.np.community.playstation.net/uploader/sec/download_data` | message attachments: np-uploader rejecting a junk ticket (result 02) |
| PS3 | `static-resource…/np/resource/ubstorage/ubstorage.xml` | online save storage: the signed discovery file (np-ubs) |
| PS4 | `dus01.ps4.update.playstation.net/update/ps4/list/us/ps4-updatelist.xml` | system update checks: ps4-update's "nothing newer" list (version 00.000.000) |
| Original Xbox | `145.241.228.207:3074` (UDP), `88/udp` on the primary | gateway and Kerberos KDC accept datagrams |
| Original Xbox | `macs.xboxlive.com` A query to xb-dns on the primary ZeroTier address | discovery DNS answers with the auth host |
| Original Xbox | `10.212.63.196:26111/26112/26115/26117/26118/26119/26120` (TCP) | accounts, terms, friends and presence, voice, signing keys, system messages, messages: mutual-TLS APIs accept connections from the status host |

## Groups (2026-11-04)

The page and `/api/status` show these groups, in this order; openpak.org/status reads the same
API, so the two never disagree about what exists.

| Group | What is in it |
|---|---|
| On the web | the site's pages and public APIs (sign-in, registration, titles, downloads, help, app API, network profile, redirect ceiling, CA, news feed, regions, the logs viewer gate) |
| Companion sites | verse, verse Lite, kart, splat, maker, rewards, docs |
| Core services | accounts, saves, messages, leaderboards, achievements, title storage, database load, registry, rewards, parental, voice, mods, Pokemon HOME / Global Link / legality, Pakka, the logs pipeline, console DNS (a DNS query for an OpenPak name only), the four NAT check ports, the TURN relay |
| On a Switch | the services behind the Nintendo names (device auth, account, shop, penne, pctl, BCAT, share) |
| Switch games (NEX / NPLN) | one session port per title, generated from the quadlets |
| Switch games with their own servers | Epic, Innersloth, Klei, Team17, Demonware, Battle.net, Stumble Guys, Fall Guys, Photon, Unity (ugs), PlayFab, Minecraft Dungeons, Unrailed, plus the vendor UDP ports |
| On a Wii U or 3DS | account (also on the IP-literal port 4443), NASC, connection test, OpenVerse, SpotPass, shop, Splatfest results, then each game's session port |
| On a Wii or DS | NAS login, connection test, DLS, game stats, Sake, race, then the WFC ports |
| Xbox 360, PS3, PS4, original Xbox | as above |

The `*.openpak.org` names resolve to the origin in public DNS, so the Wii U/3DS and Wii/DS checks need no host
mapping. Third-party names (Switch, vendors, PlayStation) are mapped to the origin in
`openpak-status.container`; add a name there when you add a check on one.

Game session ports are generated by `cmd/gen-gatus` from the production quadlets (TCP connect, or a UDP dial
which only sees a closed port, since Gatus cannot send a payload). They are shown folded on the page unless one is down.
A service without a published port that has no HTTP route (shiro, connectivity, the xb-* mutual-TLS APIs other than
the ones listed, redis, minio, mongo, mailpit) has no external check; `gen-gatus` lists them as skipped.

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
