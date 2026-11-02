Every OpenPak service has a fixed place in one port plan, so a full stack runs on one machine
without collisions and anyone can tell what a port is for. The plan comes first; the table of what
actually runs is generated and follows it.

## Rules

1. **One block of ten per service.** `+0` HTTP, `+1` gRPC, `+2` metrics and pprof, `+3` to `+9`
   spare.
2. **Nothing below 20000 and nothing at or above 27000** is allocated by the plan. Ports a
   console or game fixes for itself are the only exceptions (listed below).
3. Every service reads its listeners from `<SERVICE>_HTTP_ADDR`, `<SERVICE>_GRPC_ADDR` and
   `<SERVICE>_METRICS_ADDR`; the plan's values are the defaults and the local-run convention.
4. **A new service takes the next free block in its range**, and the pull request that adds it
   records the allocation.
5. **Only game transport (22xxx), the fixed console ports and 443/80 are reachable from
   outside.** 20xxx, 21xxx and 23xxx stay on the internal network. Metrics and pprof are never
   public.
6. Console-facing handlers of platform-neutral services (saves, news) do not get their own
   console-facing port; the adapter the console already talks to mounts them.

## Ranges

| Range | What lives there |
| --- | --- |
| 20000–20999 | Platform-neutral and adapter services: HTTP and gRPC on the internal network. |
| 21000–21999 | Console-facing TLS backends. Consoles reach everything on 443 through a TLS router (SNI passthrough or termination), which forwards to these. Never exposed directly. |
| 22000–22999 | Game transport (UDP/TCP), reachable from outside: NEX title blocks, NPLN blocks, per-publisher backends, relay helpers. |
| 23000–23999 | Local development infrastructure, and the overflow of NEX title blocks. Where they overlap, a served title wins. |
| 24000–24999 | Per-title session ports that consoles dial directly. |

## Core blocks

| Block | Service |
| --- | --- |
| 20000 | `account` (core): HTTP 20000, gRPC 20001, metrics 20002 |
| 20010 | `website` (site, portal, mobile API): HTTP 20010, metrics 20012 |
| 20020 | `saves` |
| 20030 | `mods` |
| 20040 | `news` |
| 20050 | `nn-account` (Wii U/3DS accounts) |
| 20060 | `nn-friends` (Wii U/3DS friends, gRPC side) |
| 20070 | Switch adapter internal API |
| 20080 | `nn-boss` (SpotPass) |
| 20090 / 20100 | `nn-juxtaposition` API and UI |
| 20110 | `nn-soap` (Wii U/3DS eShop SOAP) |
| 20120 | `wfc-nas` and the other `wfc-*` services (Wii/DS) |
| 20130 | `chat` |
| 20140 | network registry |
| 20150 | `connectivity` |
| 21000 | Switch adapter, console-facing |

Game servers record their own title blocks in their repository README.

## Ports fixed by consoles and games

These are outside the plan because the console or game decides them:

| Port | Service |
| --- | --- |
| 53/udp | Console DNS |
| 80, 443/tcp | Console HTTP and HTTPS, through the TLS router |
| 10025, 10125/udp | NAT check (two servers on two different public addresses) |
| 28910, 29900, 29901, 29920/tcp and 27900, 27901/udp | Wii/DS WFC (server browser, login, profiles, stats, NAT negotiation) |
| 3074/tcp+udp | Some games' lobby and NAT probes, and the original Xbox's gateway |
| 3478/udp and 49152–65535/udp | STUN/TURN relay |

## Local development infrastructure

| Port | Service |
| --- | --- |
| 23010 | S3-compatible object store (saves, mods, news, avatars) |
| 23020 | Mail catcher for the core's email flows |
| 23030 | TLS router dashboard |

## Health endpoints

Every service answers "is it running" over HTTP, so nothing infers it from a socket:
`/healthz` (liveness) and, where a service has dependencies, `/readyz`. Game servers whose game
traffic is UDP expose `/healthz` on their block's `+2` port.
