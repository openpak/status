// A status page for people, not a monitoring dashboard.
//
// Gatus does the checking — scheduling, retries, history, the arithmetic — and
// stays on localhost. This renders what it found in the site's own colours and
// in words that mean something to someone whose console will not link. The
// split is deliberate: none of the hard parts of monitoring are reimplemented
// here, and none of the presentation is left to a tool that does not know what
// any of these services are for.
package main

import (
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/openpak/status/shared"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

//go:embed icon.svg
var icon []byte

//go:embed fonts/*.woff2
var fontFS embed.FS

//go:embed page.html
var pageHTML string

//go:embed page.js
var pageJS string

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

// What Gatus reports. Only the fields this page uses.
type gatusEndpoint struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	Group   string `json:"group"`
	Results []struct {
		Success   bool      `json:"success"`
		Status    int       `json:"status"`
		Duration  int64     `json:"duration"` // nanoseconds
		Timestamp time.Time `json:"timestamp"`
	} `json:"results"`
}

// blurb is what each service means to a person. Gatus knows a name and a URL;
// it does not know that device auth failing means nobody can sign in on a
// console, and that is the only thing a reader actually wants told.
var blurb = map[string]string{
	"public_website":           "The site itself: signing in, your friends, your account.",
	"public_console-link-page": "The page your phone opens when you link a console.",
	"console_device-auth":      "The first thing a console does. If this is down, no console can reach the network at all.",
	"console_account-service":  "How a console proves who you are. Sign-in and friends on the console depend on it.",
	"console_sign-in-page":     "The sign-in screen a console shows while linking an account.",
	"console_licence-service":  "Checked while a console adds an account. If it fails, adding an account stops part-way.",
	"ps3_leaderboards":         "Posting scores and reading leaderboards in PS3 games.",
	"ps3_trophies":             "Syncing and viewing PS3 trophies.",
	"ps3_title-user-storage":   "Data a PS3 game keeps on the network for each player.",
	"ps3_xmb-feeds":            "What's New and the game and video billboards on the PS3 home screen.",
	"ps3_sign-in":              "Signing a PS3 in to the network.",
	"ps3_profiles":             "Your name and avatar as other PS3 players see them.",
	"ps3_friends":              "Your PS3 friends list.",
	"ps3_presence":             "Seeing which PS3 friends are online and what they are playing.",
	"ps3_game-update-checks":   "Checking for game updates when a PS3 game starts.",
	"ps3_message-attachments":  "Files attached to PS3 messages between friends.",
	"ps3_online-save-storage":  "Keeping PS3 game saves on the network.",
	"ps4_system-update-checks": "Checking for system software updates when a PS4 starts. OpenPak answers that nothing is newer, so consoles stay on their current firmware.",
}

// more describes the checks added with the grouped page, by group and name; keyOf makes the same
// key Gatus does (lower case, "-" for spaces and punctuation).
var more = []struct{ group, name, text string }{
	{"Public", "Sign-in page", "The page you sign in on."},
	{"Public", "Account registration", "Creating an OpenPak account."},
	{"Public", "Titles catalogue", "The list of games and what works in each, read by the website, the app and the emulators."},
	{"Public", "Downloads", "The downloads page: patchers, the app and the emulators."},
	{"Public", "Help centre", "Help articles and support tickets."},
	{"Public", "App API", "The address the phone app and the emulators talk to."},
	{"Public", "Network profile", "What the app and the console patchers fetch to learn which hosts to redirect."},
	{"Public", "Redirect ceiling", "The signed list of names a console may be redirected for. If it fails to load, console DNS drops whole families of games at its next restart."},
	{"Public", "Network CA certificate", "The certificate consoles are asked to trust."},
	{"Public", "Companion data", "Per-game companion information for the app."},
	{"Public", "News feed", "OpenPak news, as a feed."},
	{"Public", "Regions list", "Relay regions run by volunteers."},
	{"Public", "Logs viewer", "The admin log viewer, answering with its sign-in redirect."},
	{"Sites", "OpenVerse", "The Miiverse replacement for Wii U and 3DS, on the web."},
	{"Sites", "OpenVerse Lite", "Plain-HTTP OpenVerse for the PS3, Wii and DSi browsers."},
	{"Sites", "Open Kart", "Mario Kart rankings and tournaments."},
	{"Sites", "SplatNet", "Splatoon rotations and gatherings."},
	{"Sites", "Mario Maker courses", "Browsing and sharing Mario Maker courses."},
	{"Sites", "Rewards", "The rewards site: badges, themes and the shop."},
	{"Sites", "Documentation", "docs.openpak.org."},
	{"Core", "Accounts", "Creating an account, signing in, and everything that needs your identity."},
	{"Core", "Cloud saves", "Backing up and restoring game saves from an emulator or the website."},
	{"Core", "Messages", "Messages between friends on the website, the phone app and Miiverse."},
	{"Core", "Game leaderboards", "The shared score boards every console's games post to and read from."},
	{"Core", "Game achievements", "The trophies and achievements every console earns, kept in one place across consoles."},
	{"Core", "Game title storage", "The shared place where every console's games keep each player's own data, like progress counters and uploads."},
	{"Core", "Database connections", "How full the database server is. It stops accepting new connections at its limit."},
	{"Core", "Network registry", "Where the console and emulator builds fetch their signed bundles from."},
	{"Core", "Rewards", "Badges, themes and shop items earned across the network."},
	{"Core", "Parental controls", "Supervision, profiles and the OpenVerse level for children's accounts, on every console."},
	{"Core", "Voice chat", "Voice sessions for games that offer them."},
	{"Core", "Mod catalogue", "The catalogue of game mods."},
	{"Core", "Pokemon HOME", "Account-bound boxes, Pokedex, trading and the Wonder Box."},
	{"Core", "Pokemon Global Link", "The GTS, battle records and the Global Showcase."},
	{"Core", "Pokemon legality check", "Checks Pokemon before they are traded."},
	{"Core", "Pakka assistant", "The Pakka chat assistant on the website and in the app."},
	{"Core", "Logs pipeline", "Service logs reaching the admin log viewer. Fails when no log line has arrived for 15 minutes."},
	{"Core", "Console DNS", "The DNS server a console's DNS setting points at."},
	{"Core", "NAT check, first server (10025)", "Part of the NAT check that decides which online modes a console may use."},
	{"Core", "NAT check, first server (10125)", "Part of the NAT check that decides which online modes a console may use."},
	{"Core", "NAT check, second server (10025)", "Part of the NAT check that decides which online modes a console may use."},
	{"Core", "NAT check, second server (10125)", "Part of the NAT check that decides which online modes a console may use."},
	{"Core", "Relay (TURN)", "The relay that carries game traffic between players who cannot reach each other directly."},
	{"Console", "eShop", "The shop front page on the console."},
	{"Console", "Game downloads", "Where the eShop sends game and update downloads."},
	{"Console", "Online play and friends", "The connection the console keeps open for friends, presence and invitations."},
	{"Console", "Parental controls (console)", "The Parental Controls app on the console."},
	{"Console", "News (BCAT)", "News and daily content pushed to the console."},
	{"Console", "Album sharing", "Sharing album screenshots and clips to OpenVerse."},
	{"Switch vendor", "Epic Online Services", "Sign-in and matchmaking for games on Epic's online services."},
	{"Switch vendor", "Innersloth backend", "Among Us accounts."},
	{"Switch vendor", "Among Us matchmaker", "Finding and creating Among Us games."},
	{"Switch vendor", "Klei login", "Don't Starve Together sign-in."},
	{"Switch vendor", "Klei lobby", "Don't Starve Together server list."},
	{"Switch vendor", "Team17 directory", "Moving Out 2 and other Team17 games: where to find the game servers."},
	{"Switch vendor", "Demonware", "Sign-in for games on Demonware."},
	{"Switch vendor", "Battle.net region service", "Diablo II: Resurrected finds its region here."},
	{"Switch vendor", "Battle.net account", "Diablo II: Resurrected sign-in."},
	{"Switch vendor", "Stumble Guys", "Stumble Guys accounts and matchmaking."},
	{"Switch vendor", "Fall Guys login", "Fall Guys sign-in."},
	{"Switch vendor", "Photon name server", "Finds the right Photon room server for games built on Photon."},
	{"Switch vendor", "Minecraft Dungeons", "Minecraft Dungeons online services."},
	{"Switch vendor", "PlayFab (Astroneer)", "Astroneer sign-in."},
	{"Switch vendor", "Unrailed", "Unrailed online play."},
	{"Switch vendor", "Hello Kitty Island Adventure", "Hello Kitty Island Adventure invite codes and cloud saves."},
	{"Switch vendor", "Plants vs. Zombies (EA sign-in)", "Plants vs. Zombies: Battle for Neighborville sign-in."},
	{"Wii U/3DS", "Account service", "Nintendo Network ID sign-in and account setup."},
	{"Wii U/3DS", "Account terms", "The terms shown while creating or linking an account."},
	{"Wii U/3DS", "Account service (Wii U, port 4443)", "The same account service, reached the way a patched Wii U reaches it: by address, on port 4443."},
	{"Wii U/3DS", "3DS game server list", "The 3DS asks this first, to find the game server."},
	{"Wii U/3DS", "Connection test", "The page a Nintendo console loads to check that its Wi-Fi works."},
	{"Wii U/3DS", "Miiverse (API)", "OpenVerse posts and communities on a Wii U or 3DS."},
	{"Wii U/3DS", "Miiverse (Wii U web)", "The OpenVerse page on a Wii U."},
	{"Wii U/3DS", "Miiverse (3DS web)", "The OpenVerse page on a 3DS."},
	{"Wii U/3DS", "SpotPass", "Data a console downloads in the background: news, game content and StreetPass."},
	{"Wii U/3DS", "eShop", "The Wii U and 3DS shop."},
	{"Wii U/3DS", "Splatfest results", "Splatoon Splatfest scores."},
	{"Wii/DS", "DS login", "A Nintendo DS signing in to Wi-Fi Connection."},
	{"Wii/DS", "Wii login", "A Wii signing in to Wi-Fi Connection."},
	{"Wii/DS", "Connection test (Wii)", "The page a Wii loads to check that its Wi-Fi works."},
	{"Wii/DS", "Game downloads (DLS)", "Downloadable content and Wii Message Board items."},
	{"Wii/DS", "Game stats", "Game statistics and rankings."},
	{"Wii/DS", "Online data storage (Sake)", "Per-player data games keep on the network."},
	{"Wii/DS", "Race rankings", "Mario Kart Wii ghost and ranking data."},
	{"Xbox 360", "Connectivity test", "The page an Xbox 360 loads to check it is online."},
	{"Xbox 360", "Service status notice", "The service status message shown in the dashboard."},
	{"Xbox 360", "Dashboard home", "The OpenPak Home screen on the dashboard."},
	{"Xbox 360", "Live configuration", "The settings an Xbox 360 downloads when it signs in."},
	{"Xbox 360", "Token service", "Gives a signed-in Xbox 360 the tokens other services ask for."},
	{"Xbox 360", "Gamer profile (HTTPS)", "Gamer profile requests made over HTTPS."},
}

func keyOf(group, name string) string {
	r := strings.NewReplacer("/", "-", "_", "-", ",", "-", ".", "-", " ", "-", "#", "-", "+", "-", "&", "-")
	return strings.ToLower(r.Replace(group)) + "_" + strings.ToLower(r.Replace(name))
}

func init() {
	for _, m := range more {
		if _, taken := blurb[keyOf(m.group, m.name)]; !taken {
			blurb[keyOf(m.group, m.name)] = m.text
		}
	}
}

// region is a federation node as openpak.org lists it: approved, announced within the last
// few minutes, and carrying whatever relays it runs. A node that stops announcing drops off
// the list rather than turning red, which is the honest shape: nothing is claimed about it.
type region struct {
	Name, Country, Services string
	Players                 int
}

var tmpl = template.Must(template.New("page").Funcs(template.FuncMap{
	// bars is the strip: one flag per check, oldest first.
	"bars": func(h string) []bool {
		out := make([]bool, len(h))
		for i := range h {
			out[i] = h[i] == '1'
		}
		return out
	},
	"script": func() template.JS { return template.JS(pageJS) },
	"css":    func() template.CSS { return template.CSS(shared.CSS) },
	"icons":  func() template.HTML { return template.HTML(shared.Icons) },
}).Parse(pageHTML))

// cache keeps a refresh storm off Gatus: the checks run on their own schedule,
// so asking it more often than that tells nobody anything new.
type cache struct {
	mu   sync.Mutex
	at   time.Time
	view *pageView
}

func main() {
	gatus := getenv("GATUS_URL", "http://127.0.0.1:8450")
	regionsURL = getenv("OPENPAK_REGIONS_URL", "https://openpak.org/api/v1/regions")
	addr := getenv("STATUS_ADDR", "127.0.0.1:8451")
	c := &cache{}
	client := &http.Client{Timeout: 10 * time.Second}

	mux := http.NewServeMux()
	// The same rounded UI face openpak.org uses, self-hosted so this page needs nothing else.
	mux.HandleFunc("GET /fonts/{name}", func(w http.ResponseWriter, r *http.Request) {
		b, err := fontFS.ReadFile("fonts/" + r.PathValue("name"))
		if err != nil || !strings.HasSuffix(r.PathValue("name"), ".woff2") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "font/woff2")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		_, _ = w.Write(b)
	})
	mux.HandleFunc("GET /icon.svg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		_, _ = w.Write(icon)
	})
	// A machine-readable answer for anything that wants one, including the
	// app one day: the same view the page renders.
	mux.HandleFunc("GET /api/status", func(w http.ResponseWriter, r *http.Request) {
		view, err := c.get(client, gatus)
		if err != nil {
			http.Error(w, `{"error":"upstream unavailable"}`, http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(view)
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		view, err := c.get(client, gatus)
		if err != nil {
			// The checker being unreachable is itself worth saying plainly,
			// rather than showing a page that implies everything is fine.
			log.Printf("status: upstream: %v", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			view = &pageView{
				Headline: "Status is unavailable",
				Sub:      "The service that runs these checks could not be reached. This says nothing about whether OpenPak itself is working.",
				State:    "major",
			}
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, view); err != nil {
			log.Printf("status: render: %v", err)
		}
	})

	log.Printf("[status] listening %s, reading %s", addr, gatus)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func (c *cache) get(client *http.Client, gatus string) (*pageView, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.view != nil && time.Since(c.at) < 15*time.Second {
		return c.view, nil
	}
	resp, err := client.Get(gatus + "/api/v1/endpoints/statuses")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gatus: %s", resp.Status)
	}
	var eps []gatusEndpoint
	if err := json.NewDecoder(resp.Body).Decode(&eps); err != nil {
		return nil, err
	}
	c.view, c.at = build(eps), time.Now()
	addRegions(c.view, fetchRegions(client))
	return c.view, nil
}

var regionsURL string

// fetchRegions reads the public region list. Its failure is not the network's: an empty list
// draws no section, and the page says nothing it cannot back.
func fetchRegions(client *http.Client) []region {
	if regionsURL == "" {
		return nil
	}
	resp, err := client.Get(regionsURL)
	if err != nil {
		log.Printf("status: regions: %v", err)
		return nil
	}
	defer resp.Body.Close()
	var out struct {
		Regions []struct {
			Name     string            `json:"name"`
			Country  string            `json:"country"`
			Services map[string]string `json:"services"`
			Players  int               `json:"players"`
		} `json:"regions"`
	}
	if resp.StatusCode != http.StatusOK || json.NewDecoder(resp.Body).Decode(&out) != nil {
		return nil
	}
	var rs []region
	for _, r := range out.Regions {
		keys := make([]string, 0, len(r.Services))
		for k := range r.Services {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		rs = append(rs, region{Name: r.Name, Country: r.Country, Services: strings.Join(keys, ", "), Players: r.Players})
	}
	return rs
}

// addRegions lists the relay regions under Shared services, with the players each one carries.
func addRegions(v *pageView, regions []region) {
	if len(regions) == 0 {
		return
	}
	var rows []Row
	for _, r := range regions {
		rows = append(rows, Row{Name: r.Name + " (" + r.Country + ")", Note: r.Services, Up: true, Players: r.Players, HasPlayers: true})
	}
	c := Category{ID: "shared/relay-regions", Title: "Relay regions", Rows: rows, Rollup: rollup(len(rows), len(rows), "region")}
	for i := range v.Networks {
		if v.Networks[i].ID == "shared" {
			n := &v.Networks[i]
			n.Cats = append(n.Cats, c)
			n.Rollup = rollup(n.Up+len(rows), n.Total+len(rows), "check")
			return
		}
	}
}

// barCount is how much history the strip shows. Enough to see a pattern,
// few enough to stay legible on a phone.
const barCount = 40

// generatedName matches the endpoints cmd/gen-gatus adds ("wfc-gpcm 29900/tcp", "switch-arms secure 22141/tcp").
var generatedName = regexp.MustCompile(`^(.+?)( secure)? \d+/(tcp|udp)$`)

// friendly names the units whose slug does not read well; everything else is title-cased.
var friendly = map[string]string{
	"wfc-gamestats": "Game stats", "wfc-gpcm": "Friends and profiles", "wfc-gpsp": "Friend search",
	"wfc-natneg": "Connection setup", "wfc-qr2": "Matchmaking", "wfc-serverbrowser": "Server browser",
	"nn-friends": "Friends", "wiiu-chat": "Wii U chat", "np-matching2-recorder": "Matchmaking",
	"xb-auth": "Sign-in (Kerberos)", "np-community": "Friends", "among-us": "Among Us",
	"fall-guys-round": "Fall Guys rounds", "spelunky2": "Spelunky 2", "photon": "Photon", "playfab": "PlayFab",
}

var words = map[string]string{
	"3ds": "3DS", "wiiu": "Wii U", "tmnt": "TMNT", "f1": "F1", "ii": "II", "of": "of", "the": "the", "to": "to",
	"vs": "vs", "and": "and", "a": "a", "in": "in", "iv": "IV", "xx": "XX", "hff": "HFF", "mk8d": "MK8D",
}

// prettify turns a unit slug into a name a player would recognise.
func prettify(slug string) string {
	if n, ok := friendly[slug]; ok {
		return n
	}
	slug = strings.TrimPrefix(slug, "switch-")
	parts := strings.Split(slug, "-")
	for i, w := range parts {
		if r, ok := words[w]; ok {
			parts[i] = r
		} else if w != "" {
			parts[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(parts, " ")
}

func humanSince(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%d min ago", int(d.Minutes()))
	default:
		return t.Local().Format("15:04")
	}
}
