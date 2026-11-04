package main

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestWritePreview renders the page from a saved /api/status answer (PREVIEW_API) into
// PREVIEW_DIR/status-ok.html and status-fail.html, the second with a few checks failing, so a
// person can look at both. It runs only when PREVIEW_DIR is set.
func TestWritePreview(t *testing.T) {
	dir, api := os.Getenv("PREVIEW_DIR"), os.Getenv("PREVIEW_API")
	if dir == "" || api == "" {
		t.Skip("set PREVIEW_DIR and PREVIEW_API")
	}
	raw, err := os.ReadFile(api)
	if err != nil {
		t.Fatal(err)
	}
	var in struct {
		Groups []struct {
			Title           string
			Services, Ports []struct{ Name string }
		}
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		t.Fatal(err)
	}
	keyOfTitle := map[string]string{"On the web": "Public", "Companion sites": "Sites", "Core services": "Core", "On a Switch": "Console",
		"Switch games (NEX)": "Switch NEX", "Switch games (NPLN)": "Switch NPLN", "Switch games with their own servers": "Switch vendor",
		"On a Wii U or 3DS": "Wii U/3DS", "On a Wii or DS": "Wii/DS", "On an Xbox 360": "Xbox 360", "On a PS3": "PS3", "On a PS4": "PS4",
		"On an original Xbox": "Original Xbox"}
	mk := func(failing map[string]bool) []gatusEndpoint {
		rng := rand.New(rand.NewSource(7))
		var eps []gatusEndpoint
		add := func(group, name string) {
			e := gatusEndpoint{Key: keyOf(group, name), Name: name, Group: group}
			now := time.Now()
			for i := 0; i < 60; i++ {
				ok := rng.Intn(60) != 0
				if i == 59 {
					ok = !failing[name]
				}
				e.Results = append(e.Results, struct {
					Success   bool      `json:"success"`
					Status    int       `json:"status"`
					Duration  int64     `json:"duration"`
					Timestamp time.Time `json:"timestamp"`
				}{ok, 200, int64(10+rng.Intn(80)) * int64(time.Millisecond), now.Add(-time.Duration(59-i) * 2 * time.Minute)})
			}
			eps = append(eps, e)
		}
		for _, g := range in.Groups {
			for _, s := range g.Services {
				add(keyOfTitle[g.Title], s.Name)
			}
			for _, p := range g.Ports {
				add(keyOfTitle[g.Title], strings.ReplaceAll(strings.ToLower(p.Name), " ", "-")+" 22141/tcp")
			}
		}
		return eps
	}
	for name, failing := range map[string]map[string]bool{"ok": {}, "fail": {"eShop": true, "Arms": true, "Friends and presence": true}} {
		f, err := os.Create(filepath.Join(dir, "status-"+name+".html"))
		if err != nil {
			t.Fatal(err)
		}
		if err := tmpl.Execute(f, build(mk(failing))); err != nil {
			t.Fatal(err)
		}
		f.Close()
	}
}
