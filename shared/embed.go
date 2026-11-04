// Package shared is the status pages' one visual language: the badge, the card, the
// uptime strip and the icons. status.openpak.org embeds it; openpak.org/status gets a
// byte-for-byte copy through scripts/sync-shared.sh (the website's CI builds alone, so
// it cannot import this repository), and the website's tests fail when the copy drifts.
package shared

import _ "embed"

//go:embed statusview.css
var CSS string

//go:embed icons.html
var Icons string
