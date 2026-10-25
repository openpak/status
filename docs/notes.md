# Notes: status

The network's status page. It runs on **Draughts**, deliberately not on the
production box: a status page that lives on the machine it watches reports "up"
right until the moment it cannot report at all, which is the only moment anyone
reads it.

## Why two processes

Gatus does the monitoring and stays private; `cmd/statuspage` renders what it
found in the site's own colours and words. None of the hard parts of monitoring
are reimplemented, and none of the presentation is left to a tool that does not
know what any of these services are for.
