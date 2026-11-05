# dedup-check

`openpak-dedup-check` answers one question nightly on the main box: do the caches and adapter copies still agree with
the account core? Board card openpak/prds#131 phase 10, guard against player-data duplication coming back. It is
read-only (SELECT through `podman exec openpak-postgres psql`) and prints counts only: every comparison is an md5 of the
lowercased, trimmed value computed inside Postgres and joined on the core account id.

| File | Goes to |
| --- | --- |
| `openpak-dedup-check` | `~/.local/bin/openpak-dedup-check` |
| `openpak-dedup-check.service` / `.timer` | `~/.config/systemd/user/` (daily 07:45 UTC; `OnFailure=openpak-unit-failed@%n`) |

Stores compared with core `accounts` / `account_miis` / `play_sessions`: nn_account `pnids`; nn_friends caches;
miiverse `users`; xbox `xb_account.users` and `gamertag_aliases`; website `web_sessions` (ids only); nxbaas `baas.json`
users and `accounts.json by_account`; spelunky2 runs; rewards `accounts` and `play_cursor`. Photon, stumble, spelunky2,
wwfc and the website have no core link column, so a schema guard proves no identity column (name, email, country,
birthday, language, timezone, Mii) came back there.

Documented exceptions, each with evidence in `CHANGELOG.md` v0.7.1: rewards compares with the core's wall-clock play history (what rewards reads), retired Xbox users (`unlinked`/`deleted`) are skipped (`retired-skipped`), and an unlinked nx-baas user younger than 7 days is `grace`, not an orphan.

Mii caches are compared by md5 of the Ver3 bytes against the core's Ver3 (the "Default" placeholder nn-account serves counts as no Mii); `pnids` holds none, and `nn_account.pending_miis` (the core-outage upload queue) fails on a row older than an hour. Miiverse's official Pakka account is `service-skipped`.

Per store: `compared`, `equal`, `diverging`, `orphans` (rows naming an account the core does not have). A row diverges
when a cache holds a non-empty value that differs from the core, or a value the core lacks (`(core empty)`); an empty
cache value is fine. The run fails (exit 1, `OnFailure=` makes Pakka announce it) on any divergence, any orphan count
above the last run's (`~/openpak/deploy/dedup-state.json`), a store that cannot be read, or a new identity column.

Output: one line per run to stdout (journal, then openpak-logd writes it to
`~/openpak-logs/openpak-dedup-check/<day>.log`, pruned after 90 days); detail in `~/openpak/deploy/dedup-report.txt`.
By hand: `openpak-dedup-check --verbose` (about 10 seconds; `DEDUP_NO_STATE=1` leaves the orphan baseline alone).

The CI twin for new adapter tables and columns is `scripts/dedup-lint.py` in the public `openpak/.github` repo (reusable workflow `dedup-lint.yml`, per-repo `.dedup-allow`).
