# Changelog — status

Generated from git history on 2026-09-15. `git log` stays the source
of truth; this file is the readable summary.

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

