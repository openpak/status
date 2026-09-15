# How federation works

> Trimmed copy for this repository; the canonical document lives at
> `Openpak/docs/federation.md` and governs. Last synchronised 2026-09-15.

Someone in Australia runs a small box, registers it, and from then on rooms hosted by
Australians are served from Australia. Their friends in France join those rooms the same way
they always did. Nobody in Australia has an Australian account, an Australian friends list or
an Australian matchmaking server: there is one of each, at openpak.org, and the box in Sydney
is only where the packets go.

…

## The parts

…

**The website** is the node's only door. `POST /api/v1/federation/announce` takes the node's
own key as a bearer and relays it to the core, so the core's gRPC never leaves the internal
network. `GET /api/v1/regions` is the public list of approved nodes seen in the last three
minutes, with their services and live player counts, and it is the one thing every
matchmaking server, the status page and the app read. The admin page mints a node's key
(shown once, the core keeps a hash) and flips approval.

…

## The announce cycle

Once a minute the node posts what it runs:

```json
{"services":{"photon":"203.0.113.9:22002","turn":"203.0.113.9:3478","stun":"203.0.113.9:3478"}}
```

The core replaces the node's `services` and touches `last_seen_at`. A node that misses three
announces drops out of `/regions`; the matchmaking servers see that within another minute and
stop placing rooms there. Rooms already on it keep going until they end: nothing tears them
down, and nothing needs to.

A service a node does not run is left out, and home's own is used for it. A node can be
TURN-only, or Photon-only.
