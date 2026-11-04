package main

import (
	"strings"
	"testing"
	"time"
)

func at(success ...bool) gatusEndpoint {
	e := gatusEndpoint{Key: "public_website", Name: "Website", Group: "Public"}
	for _, s := range success {
		e.Results = append(e.Results, struct {
			Success   bool      `json:"success"`
			Status    int       `json:"status"`
			Duration  int64     `json:"duration"`
			Timestamp time.Time `json:"timestamp"`
		}{Success: s, Duration: int64(12 * time.Millisecond), Timestamp: time.Now()})
	}
	return e
}

// The headline is the whole page for most readers, so it has to be right for
// each of the three cases rather than only the happy one.
func TestHeadlineReflectsWhatIsActuallyDown(t *testing.T) {
	up, down := at(true, true, true), at(false)
	down.Key, down.Name = "console_device-auth", "Device auth"
	down.Group = "Console"

	if v := build([]gatusEndpoint{up}); v.State != "ok" {
		t.Errorf("all up: %q %q", v.State, v.Headline)
	}
	if v := build([]gatusEndpoint{up, down}); v.State != "some" {
		t.Errorf("one down: %q %q", v.State, v.Headline)
	}
	if v := build([]gatusEndpoint{down}); v.State != "major" {
		t.Errorf("all down: %q %q", v.State, v.Headline)
	}
	// Nothing checked is not the same as everything fine, and saying so
	// wrongly is the one failure a status page cannot have.
	if v := build(nil); v.State != "major" {
		t.Errorf("no data must not read as healthy: %q %q", v.State, v.Headline)
	}
}

func TestUptimeAndHistory(t *testing.T) {
	e := at(true, false, true, true) // 3 of 4
	v := build([]gatusEndpoint{e})
	s := v.Networks[0].Cats[0].Rows[0]
	if s.Uptime < 74.9 || s.Uptime > 75.1 {
		t.Errorf("uptime = %v, want 75", s.Uptime)
	}
	if s.Hist != "1011" {
		t.Errorf("history strip wrong: %q", s.Hist)
	}
	if !s.Up {
		t.Error("the latest result decides the current state")
	}
	// Every service the page shows should explain itself; an unknown key is
	// allowed to have no blurb, but the ones we ship should not.
	if s.Note == "" {
		t.Error("no blurb for a known service")
	}
}

// A long-running endpoint keeps more results than the strip shows; the strip
// must show the most recent, not the oldest.
func TestHistoryStripShowsTheEnd(t *testing.T) {
	states := make([]bool, barCount+10)
	for i := range states {
		states[i] = true
	}
	states[0] = false // an old failure, outside the window
	v := build([]gatusEndpoint{at(states...)})
	if h := v.Networks[0].Cats[0].Rows[0].Hist; len(h) != barCount || strings.Contains(h, "0") {
		t.Fatal("the strip is showing results older than the window")
	}
}

// Each level shows the worst beneath it, and only the failing path starts open.
func TestRollupAndOpenPath(t *testing.T) {
	ok := at(true)
	ok.Group, ok.Name, ok.Key = "Console", "eShop", "console_eshop"
	bad := at(false)
	bad.Group, bad.Name, bad.Key = "Switch NEX", "Arms 22141/tcp", "x"
	v := build([]gatusEndpoint{ok, bad, at(true)})
	var sw, web *Network
	for i := range v.Networks {
		switch v.Networks[i].ID {
		case "switch":
			sw = &v.Networks[i]
		case "websites":
			web = &v.Networks[i]
		}
	}
	if sw == nil || web == nil {
		t.Fatalf("networks: %+v", v.Networks)
	}
	if sw.State != "partial" || !sw.Open || web.Open {
		t.Errorf("switch %q open=%v, websites open=%v", sw.State, sw.Open, web.Open)
	}
	for _, c := range sw.Cats {
		if (c.Title == "Services") == c.Open {
			t.Errorf("category %q open=%v", c.Title, c.Open)
		}
	}
	if len(v.Groups) != 3 {
		t.Errorf("legacy groups kept for the network map: %+v", v.Groups)
	}
	if got := groupIssues([]Issue{{"Switch", "eShop", false}, {"Switch", "ARMS", true}, {"Wii U / 3DS", "eShop", false}}, 3); got != "Switch: eShop and ARMS; Wii U / 3DS: eShop" {
		t.Errorf("grouped: %q", got)
	}
	if len(v.Issues) != 1 || !v.Issues[0].Game {
		t.Errorf("issues: %+v", v.Issues)
	}
	if len(v.Problems) != 1 || v.Problems[0] != "Switch: Arms" {
		t.Errorf("problems: %v", v.Problems)
	}
}

// Networks read family by family, oldest first, with Shared services and Websites last; a network
// nobody runs (PSP) has a place but does not appear.
func TestNetworkOrder(t *testing.T) {
	mk := func(group, name string) gatusEndpoint {
		e := at(true)
		e.Group, e.Name, e.Key = group, name, keyOf(group, name)
		return e
	}
	v := build([]gatusEndpoint{mk("PS4", "System update checks"), mk("Original Xbox", "Accounts"), mk("Xbox 360", "Dashboard home"),
		mk("Console", "eShop"), mk("Wii U/3DS", "eShop"), mk("Wii/DS", "DS login"), mk("PS3", "Sign-in"), mk("Core", "Accounts"), mk("Public", "Website")})
	var ids []string
	for _, n := range v.Networks {
		ids = append(ids, n.ID)
	}
	if got, want := strings.Join(ids, ","), "wii-ds,wiiu-3ds,switch,ps3,ps4,xbox,xbox360,shared,websites"; got != want {
		t.Errorf("order %s, want %s", got, want)
	}
	if !v.Networks[0].NewFamily || v.Networks[1].NewFamily || !v.Networks[3].NewFamily || !v.Networks[5].NewFamily || !v.Networks[7].NewFamily || v.Networks[8].NewFamily {
		t.Errorf("family headings: %+v", v.Networks)
	}
}
