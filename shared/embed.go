// Package shared is the status pages' one visual language: the badge, the card, the
// uptime strip and the icons. status.openpak.org embeds it; openpak.org/status gets a
// byte-for-byte copy through scripts/sync-shared.sh (the website's CI builds alone, so
// it cannot import this repository), and the website's tests fail when the copy drifts.
package shared

import (
	_ "embed"
	"encoding/json"
)

//go:embed statusview.css
var CSS string

//go:embed icons.html
var Icons string

//go:embed networks.json
var NetworksJSON []byte

// Net is one entry of the single network order: family, then generation, oldest first, then
// Shared services and Websites. Networks nobody runs yet (PSP, PS Vita, Xbox One) are defined here
// so the order is settled, and appear on a page only once they have something to show.
type Net struct {
	ID, Title, Family string
	Consoles          []string
	Site              string
}

// Networks is the order, read once.
var Networks = func() []Net {
	var n []Net
	if err := json.Unmarshal(NetworksJSON, &n); err != nil {
		panic("shared/networks.json: " + err.Error())
	}
	return n
}()
