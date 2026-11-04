package main

import (
	"fmt"
	"github.com/openpak/status/shared"
	"regexp"
	"sort"
	"strings"
	"time"
)

// The page is three levels deep: a network (or Shared services, or Websites), the categories
// inside it, and the individual checks. Every level carries the worst state beneath it.

// Row is one check. Players are not known here (this page must keep answering when the main
// site cannot), so the row carries history and uptime instead.
type Row struct {
	Name, Note string
	Up         bool
	Uptime     float64
	Latency    string
	Checked    string
	Hist       string // oldest first, "1" answered and "0" did not; drawn as the strip
	HistEnd    int64  // unix seconds of the newest result
	HistStep   int    // seconds between results
	Players    int
	HasPlayers bool
}

// Rollup is what a category or network says about the rows beneath it.
type Rollup struct {
	State string // ok | partial | down
	Label string // Operational | Partly down | Down
	Up    int
	Total int
	Sum   string // "40 services working" or "2 of 40 not working"
	Open  bool   // something beneath is failing, so the path to it starts open
}

type Category struct {
	ID, Title string
	Rows      []Row
	Rollup
}

type Network struct {
	ID, Title, Note, Family string
	NewFamily               bool // the first network shown of its family, or of the shared group (empty family): the page draws a heading or a gap above it
	Cats                    []Category
	Rollup
}

type pageView struct {
	Headline, Sub string
	State         string // ok | some | major
	Networks      []Network
	Groups        []legacyGroup
	Problems      []string // "Switch: eShop" ... for anything that has the website's attention
	Issues        []Issue  // the same, structured, so openpak.org can group them and use its own game names
	Down, Total   int
	Generated     string
	Window        string
	Search        bool
}

// legacyGroup keeps the shape /api/status had before the drill-down (a group per Gatus group,
// services and session ports apart): openpak.org's network map reads node states from it.
type legacyGroup struct {
	Title, Note     string
	Services, Ports []legacyRow
}

type legacyRow struct {
	Name   string
	Up     bool
	Uptime float64
}

var legacyTitles = []struct{ Key, Title string }{
	{"Public", "On the web"}, {"Sites", "Companion sites"}, {"Core", "Core services"}, {"Console", "On a Switch"},
	{"Switch NEX", "Switch games (NEX)"}, {"Switch NPLN", "Switch games (NPLN)"}, {"Switch vendor", "Switch games with their own servers"},
	{"Wii U/3DS", "On a Wii U or 3DS"}, {"Wii/DS", "On a Wii or DS"}, {"Xbox 360", "On an Xbox 360"}, {"PS3", "On a PS3"},
	{"PS4", "On a PS4"}, {"Original Xbox", "On an original Xbox"},
}

// Issue is one failing check: its network's title, its name, and whether it is a game server.
type Issue struct {
	Network, Name string
	Game          bool
}

// groupIssues reads failing checks as "Switch: eShop and Arms; Wii U / 3DS: eShop", each network once.
func groupIssues(is []Issue, perNet int) string {
	var order []string
	by := map[string][]string{}
	for _, i := range is {
		if _, ok := by[i.Network]; !ok {
			order = append(order, i.Network)
		}
		by[i.Network] = append(by[i.Network], i.Name)
	}
	var parts []string
	for _, n := range order {
		names := by[n]
		extra := 0
		if len(names) > perNet {
			extra, names = len(names)-perNet, names[:perNet]
		}
		var s string
		switch {
		case extra > 0:
			s = strings.Join(names, ", ") + fmt.Sprintf(" and %d more", extra)
		case len(names) > 1:
			s = strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1]
		default:
			s = names[0]
		}
		parts = append(parts, n+": "+s)
	}
	return strings.Join(parts, "; ")
}

type check struct {
	group, name string
	port        bool
	row         Row
}

var (
	reSwitchPort = regexp.MustCompile(`(?i)\bports?\b|^unity `)
	reConnect    = regexp.MustCompile(`(?i)^console dns|^nat check|^relay`)
)

// Wii U / 3DS entries that are session ports in Gatus but are services to a reader.
var wiiuServiceLike = map[string]bool{"Friends": true, "Wii U chat": true}

// place decides which network and category a check belongs to.
func place(c check) (net, cat string) {
	switch c.group {
	case "Public":
		return "websites", "Websites and app"
	case "Sites":
		return "websites", "Companion sites"
	case "Core":
		if reConnect.MatchString(c.name) {
			return "shared", "Connectivity"
		}
		return "shared", "Core services"
	case "Console":
		if reSwitchPort.MatchString(c.name) {
			return "switch", "Game servers"
		}
		return "switch", "Services"
	case "Switch NEX", "Switch NPLN":
		return "switch", "Game servers"
	case "Switch vendor":
		return "switch", "Third-party game servers"
	case "Wii U/3DS":
		if c.port && !wiiuServiceLike[c.name] {
			return "wiiu-3ds", "Game servers"
		}
		return "wiiu-3ds", "Services"
	case "Wii/DS":
		return "wii-ds", "Services"
	case "Xbox 360":
		return "xbox360", "Services"
	case "Original Xbox":
		return "xbox", "Services"
	case "PS3":
		return "ps3", "Services"
	case "PS4":
		return "ps4", "Services"
	}
	return "shared", "Core services"
}

// networkNotes is the sentence under each network's name; the order and titles come from shared/networks.json.
var networkNotes = map[string]string{
	"switch":   "Accounts, the eShop, friends and the game servers a Switch reaches.",
	"wiiu-3ds": "Accounts, OpenVerse, SpotPass, the shop and each game's session port.",
	"wii-ds":   "Wi-Fi Connection: sign-in, downloads, game stats and matchmaking.",
	"xbox360":  "Xbox Live services an Xbox 360 reaches through the OpenPak gateway.",
	"xbox":     "Sign-in, friends, accounts and voice for the original Xbox.",
	"ps3":      "PlayStation Network services a PS3 reaches.",
	"ps4":      "PlayStation Network services a PS4 reaches.",
	"shared":   "What every console network shares: accounts, saves, scores, DNS, the NAT check and the relay.",
	"websites": "What you use in a browser or on your phone, and the sites that go with the games.",
}

var catOrder = []string{"Services", "Game servers", "Third-party game servers", "Core services", "Connectivity", "Websites and app", "Companion sites", "Relay regions"}

func catID(title string) string {
	return strings.Trim(regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(strings.ToLower(title), "-"), "-")
}

func rollup(up, total int, noun string) Rollup {
	r := Rollup{Up: up, Total: total}
	switch {
	case total == 0:
		r.State, r.Label, r.Sum = "ok", "Operational", "nothing checked"
	case up == total:
		r.State, r.Label = "ok", "Operational"
		r.Sum = fmt.Sprintf("%d %s working", total, plural(total, noun))
	case up == 0:
		r.State, r.Label, r.Open = "down", "Down", true
		r.Sum = fmt.Sprintf("%d of %d not working", total-up, total)
	default:
		r.State, r.Label, r.Open = "partial", "Partly down", true
		r.Sum = fmt.Sprintf("%d of %d not working", total-up, total)
	}
	return r
}

func plural(n int, noun string) string {
	if n == 1 {
		return noun
	}
	return noun + "s"
}

// assemble groups the checks into networks and categories and rolls the states up.
func assemble(checks []check) []Network {
	type key struct{ net, cat string }
	rows := map[key][]Row{}
	for _, c := range checks {
		n, cat := place(c)
		rows[key{n, cat}] = append(rows[key{n, cat}], c.row)
	}
	var out []Network
	for _, n := range shared.Networks {
		net := Network{ID: n.ID, Title: n.Title, Note: networkNotes[n.ID], Family: n.Family}
		var up, total int
		for _, cat := range catOrder {
			rs := rows[key{n.ID, cat}]
			if len(rs) == 0 {
				continue
			}
			// A session port and a service can share a name ("Game stats"); tell them apart.
			seen := map[string]int{}
			for _, r := range rs {
				seen[r.Name]++
			}
			for i := range rs {
				if seen[rs[i].Name] > 1 && rs[i].Note == "" {
					rs[i].Name += " (session)"
				}
			}
			// Failures first, so the one a reader opened this for is at the top.
			sort.SliceStable(rs, func(i, j int) bool {
				if rs[i].Up != rs[j].Up {
					return !rs[i].Up
				}
				return rs[i].Name < rs[j].Name
			})
			c := Category{ID: n.ID + "/" + catID(cat), Title: cat, Rows: rs}
			var cu int
			for _, r := range rs {
				if r.Up {
					cu++
				}
			}
			noun := "service"
			if strings.Contains(cat, "game servers") || cat == "Game servers" {
				noun = "game server"
			}
			c.Rollup = rollup(cu, len(rs), noun)
			net.Cats = append(net.Cats, c)
			up, total = up+cu, total+len(rs)
		}
		if total == 0 {
			continue
		}
		net.Rollup = rollup(up, total, "check")
		net.NewFamily = len(out) > 0 && out[len(out)-1].Family != n.Family || len(out) == 0 && n.Family != ""
		out = append(out, net)
	}
	return out
}

func histOf(results []time.Time, ok []bool) (string, int64, int) {
	var b strings.Builder
	for _, s := range ok {
		if s {
			b.WriteByte('1')
		} else {
			b.WriteByte('0')
		}
	}
	step := 0
	if n := len(results); n > 1 {
		step = int(results[n-1].Sub(results[0]).Seconds()) / (n - 1)
	}
	var end int64
	if n := len(results); n > 0 {
		end = results[n-1].Unix()
	}
	return b.String(), end, step
}

func build(eps []gatusEndpoint) *pageView {
	var checks []check
	var down, total int
	for _, e := range eps {
		if len(e.Results) == 0 {
			continue
		}
		name, bl, isPort := e.Name, blurb[e.Key], false
		if m := generatedName.FindStringSubmatch(e.Name); m != nil {
			name, bl, isPort = prettify(m[1]), "", true
		}
		last := e.Results[len(e.Results)-1]
		total++
		if !last.Success {
			down++
		}
		results := e.Results
		if len(results) > barCount {
			results = results[len(results)-barCount:]
		}
		ts, ok := make([]time.Time, len(results)), make([]bool, len(results))
		for i, r := range results {
			ts[i], ok[i] = r.Timestamp, r.Success
		}
		var good int
		for _, r := range e.Results {
			if r.Success {
				good++
			}
		}
		h, end, step := histOf(ts, ok)
		checks = append(checks, check{group: e.Group, name: name, port: isPort, row: Row{
			Name: name, Note: bl, Up: last.Success,
			Uptime:  float64(good) / float64(len(e.Results)) * 100,
			Latency: fmt.Sprintf("%d ms", last.Duration/int64(time.Millisecond)),
			Checked: humanSince(last.Timestamp),
			Hist:    h, HistEnd: end, HistStep: step,
		}})
	}

	v := &pageView{
		Generated: time.Now().Local().Format("15:04 MST"),
		Window:    fmt.Sprintf("last %d checks", barCount),
		Down:      down, Total: total, Search: true,
	}
	v.Networks = assemble(checks)
	for _, lt := range legacyTitles {
		g := legacyGroup{Title: lt.Title}
		for _, c := range checks {
			if c.group != lt.Key {
				continue
			}
			r := legacyRow{Name: c.name, Up: c.row.Up, Uptime: c.row.Uptime}
			if c.port {
				g.Ports = append(g.Ports, r)
			} else {
				g.Services = append(g.Services, r)
			}
		}
		if len(g.Services)+len(g.Ports) > 0 {
			v.Groups = append(v.Groups, g)
		}
	}
	switch {
	case total == 0:
		v.State, v.Headline = "major", "Nothing is being checked"
		v.Sub = "No results yet."
	case down == 0:
		v.State, v.Headline = "ok", "Everything is working"
		v.Sub = fmt.Sprintf("All %d checks answered the last time they ran.", total)
	case down >= total:
		v.State, v.Headline = "major", "Everything is down"
		v.Sub = "No service answered its last check."
	default:
		v.State = "some"
		v.Headline = fmt.Sprintf("%d %s having problems", down, map[bool]string{true: "thing is", false: "things are"}[down == 1])
		v.Sub = groupIssues(v.Issues, 3)
	}
	for _, n := range v.Networks {
		for _, c := range n.Cats {
			for _, r := range c.Rows {
				if !r.Up {
					v.Issues = append(v.Issues, Issue{Network: n.Title, Name: r.Name, Game: strings.HasSuffix(c.Title, "game servers") || c.Title == "Game servers"})
					if len(v.Problems) < 8 {
						v.Problems = append(v.Problems, n.Title+": "+r.Name)
					}
				}
			}
		}
	}
	return v
}
