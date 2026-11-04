# Changelog — status

Generated from git history on 2026-09-15. `git log` stays the source
of truth; this file is the readable summary.

## v0.2.0

- The page and `/api/status` show every check, in thirteen groups: On the web, Companion sites, Core services, On a Switch, Switch games (NEX), Switch games (NPLN), Switch games with their own servers, On a Wii U or 3DS, On a Wii or DS, Xbox 360, PS3, PS4, original Xbox. The generated game-session-port checks are listed too (friendly names, one line each, folded unless one is down) and count in the headline; each group says "N of M working". `/api/status` carries them as `Ports`.
- 248 checks, up from 168. New: the public site's pages and APIs (sign-in, registration, titles, downloads, help, app API, network profile, redirect ceiling, CA, companion data, news feed, regions, logs viewer gate); companion sites (OpenVerse, Lite, Open Kart, SplatNet, Mario Maker, Rewards, docs); core services (registry, rewards, parental, voice, mods, Pokemon HOME / Global Link / legality, Pakka, logs pipeline, console DNS by a DNS query for an OpenPak name, the four NAT check ports, TURN); Switch (eShop, game downloads, penne, parental, BCAT, album sharing) and vendor backends (Epic, Innersloth, Among Us, Klei, Team17, Demonware, Battle.net, Stumble Guys, Fall Guys, Photon, Minecraft Dungeons, PlayFab, Unrailed); Wii U/3DS (account incl. the IP-literal port 4443, NASC, connection test, OpenVerse, SpotPass, shop, Splatfest); Wii/DS (DS and Wii login, DLS, game stats, Sake, race); Xbox 360 (connectivity test, service status, dashboard home, Live configuration, token service, profile).
- The account, saves, chat, leaderboards, achievements, title storage and database checks moved from Public to Core services.
- `gen-gatus` skips `np-matching2-recorder` (capture-only, started by hand).
- The container maps the extra third-party names to the origin (`PodmanArgs`).

## v0.1.24

- Console group: Unity sign-in and Unity lobby checks for Gang Beasts (ugs); the prober maps the two Unity names to the origin.

## v0.1.23

- Public group: Database connections check (`/healthz/db-connections`, website v0.86.3): 503 once Postgres holds 240 client connections, 80% of `max_connections=300`.

## v0.1.22

- Public group: the Accounts check reads `https://api.openpak.org/healthz/core`; `account.openpak.org` is now the Wii U/3DS console sign-in host, not the core.

## v0.1.21

- Public group: Game title storage check (`/healthz/title-storage`, the neutral title storage core behind every console's per-game player data).

## v0.1.20

- PS3 group: Trophies check (trophy discovery through np-account and np-trophy).

## v0.1.19

- Public group: Game achievements check (`/healthz/achievements`, the neutral achievements core behind PS3 trophies and later every console's achievements).

## v0.1.18

- Public group: Game leaderboards check (`/healthz/leaderboards`, the neutral leaderboards core behind every console's score server).

## v0.1.17

- Original Xbox group: Online storage, Player feedback and Auto-update referral checks (xb-stfd, xb-fbserver, xb-autoupd)

## v0.1.15

- PS3 group: online save storage (np-ubs)

## v0.1.14

- Original Xbox group: Signing keys, System messages and Messages checks (xb-xsig, xb-msgserver, xb-xmessaging)

## v0.1.13

- PS4 group with a system update checks probe (ps4-update's "nothing newer" list via dus01.ps4.update.playstation.net).

## v0.1.12

- Rankings API check matches the `games` key.

## v0.1.11

- Checks for the rankings hub (`/rankings`) and the public rankings API (`/api/v1/rankings`).

## v0.1.10

- Xbox 360 check "Online dashboard modules" (x360-som manifest).

## v0.1.9

- PS3 game update checks and message attachments: page blurbs and the status container's host entries
  (the checks themselves landed in v0.1.8)

## v0.1.8

- PS3 group: game update checks (np-update) and message attachments (np-uploader)

- Xbox 360 group: Messages check (Live service 18)

## v0.1.7

- Gateway check renamed so the page shows it

## v0.1.6

- Original Xbox group: gateway, discovery DNS, accounts, terms, friends and voice checks; generated xb-auth moved into it

## v0.1.5

- Xbox 360 group: Console DNS check (x360-dns)

## v0.1.4

- PS3 group: sign-in, profiles, friends and presence checks

## v0.1.3

- Xbox 360 group: accounts, presence and friends, profile text check, country information, certificate
  validation, revocation list and profile settings, one check per Live service

## v0.1.2

- PS3 group: leaderboards, Title User Storage and XMB feed checks

## v0.1.1

- Xbox 360 group: matchmaking, leaderboards and ranked-match arbitration checks

## Unreleased

- Go 1.26, the house version [71d1b3f]
- gatus: watch the Stardew and Wonder session ports on the origin IP [38c257f]
- housekeeping: daily org-wide cleanup keeping the newest 4 artifact runs and image versions
  [288a9c1]; artifacts older than 3 days deleted, and the job can check out this repo
  [a626ede]
- ci: housekeeping on self-hosted runners; build scratch in the workspace, not `/tmp`


## v0.1.0 — 2026-09-09

- Show the region nodes [9bd4363]
- Watch the default profile picture from off-box [5ad7a60]
- statuspage: the branded renderer, its unit, and a README that matches the box [ff3a1ae]
- Status page for the network, on a machine that is not the network [31f90ab]

