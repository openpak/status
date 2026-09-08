package main

import (
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
	s := v.Groups[0].Services[0]
	if s.Uptime < 74.9 || s.Uptime > 75.1 {
		t.Errorf("uptime = %v, want 75", s.Uptime)
	}
	if len(s.Bars) != 4 || s.Bars[1].Up {
		t.Errorf("history strip wrong: %+v", s.Bars)
	}
	if !s.Up {
		t.Error("the latest result decides the current state")
	}
	// Every service the page shows should explain itself; an unknown key is
	// allowed to have no blurb, but the ones we ship should not.
	if s.Blurb == "" {
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
	for _, b := range v.Groups[0].Services[0].Bars {
		if !b.Up {
			t.Fatal("the strip is showing results older than the window")
		}
	}
}
