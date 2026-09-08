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
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

//go:embed icon.svg
var icon []byte

//go:embed page.html
var pageHTML string

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
	"public_accounts":          "Creating an account, signing in, and everything that needs your identity.",
	"public_console-link-page": "The page your phone opens when you link a console.",
	"console_device-auth":      "The first thing a console does. If this is down, no console can reach the network at all.",
	"console_account-service":  "How a console proves who you are. Sign-in and friends on the console depend on it.",
	"console_sign-in-page":     "The sign-in screen a console shows while linking an account.",
	"console_licence-service":  "Checked while a console adds an account. If it fails, adding an account stops part-way.",
}

// Human names for the groups, and the order they read in.
var groups = []struct{ Key, Title, Note string }{
	{"Public", "On the web", "What you use in a browser or on your phone."},
	{"Console", "On a console", "What your Switch talks to. These are checked the way a console reaches them, not the way a browser would."},
}

type service struct {
	Name, Blurb string
	Up          bool
	Uptime      float64
	Latency     string
	Checked     string
	Bars        []bar
}

type bar struct {
	Up    bool
	Title string
}

type groupView struct {
	Title, Note string
	Services    []service
}

type pageView struct {
	Headline, Sub string
	State         string // ok | some | major
	Groups        []groupView
	Generated     string
	Window        string
}

var tmpl = template.Must(template.New("page").Parse(pageHTML))

// cache keeps a refresh storm off Gatus: the checks run on their own schedule,
// so asking it more often than that tells nobody anything new.
type cache struct {
	mu   sync.Mutex
	at   time.Time
	view *pageView
}

func main() {
	gatus := getenv("GATUS_URL", "http://127.0.0.1:8450")
	addr := getenv("STATUS_ADDR", "127.0.0.1:8451")
	c := &cache{}
	client := &http.Client{Timeout: 10 * time.Second}

	mux := http.NewServeMux()
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
	return c.view, nil
}

// barCount is how much history the strip shows. Enough to see a pattern,
// few enough to stay legible on a phone.
const barCount = 40

func build(eps []gatusEndpoint) *pageView {
	byGroup := map[string][]service{}
	var down, total int
	for _, e := range eps {
		if len(e.Results) == 0 {
			continue
		}
		last := e.Results[len(e.Results)-1]
		total++
		if !last.Success {
			down++
		}

		var ok int
		results := e.Results
		if len(results) > barCount {
			results = results[len(results)-barCount:]
		}
		bars := make([]bar, 0, len(results))
		for _, r := range results {
			bars = append(bars, bar{
				Up:    r.Success,
				Title: r.Timestamp.Local().Format("15:04") + statusWord(r.Success),
			})
		}
		for _, r := range e.Results {
			if r.Success {
				ok++
			}
		}
		byGroup[e.Group] = append(byGroup[e.Group], service{
			Name:    e.Name,
			Blurb:   blurb[e.Key],
			Up:      last.Success,
			Uptime:  float64(ok) / float64(len(e.Results)) * 100,
			Latency: fmt.Sprintf("%d ms", last.Duration/int64(time.Millisecond)),
			Checked: humanSince(last.Timestamp),
			Bars:    bars,
		})
	}

	view := &pageView{
		Generated: time.Now().Local().Format("15:04 MST"),
		Window:    fmt.Sprintf("last %d checks", barCount),
	}
	switch {
	case total == 0:
		view.State, view.Headline = "major", "Nothing is being checked"
		view.Sub = "No results yet."
	case down == 0:
		view.State, view.Headline = "ok", "Everything is working"
		view.Sub = "All services responded to their last check."
	case down >= total:
		view.State, view.Headline = "major", "Everything is down"
		view.Sub = "No service responded to its last check."
	default:
		view.State, view.Headline = "some", "Some things are having trouble"
		view.Sub = fmt.Sprintf("%d of %d services did not respond to their last check.", down, total)
	}

	for _, g := range groups {
		s := byGroup[g.Key]
		if len(s) == 0 {
			continue
		}
		sort.Slice(s, func(i, j int) bool { return s[i].Name < s[j].Name })
		view.Groups = append(view.Groups, groupView{Title: g.Title, Note: g.Note, Services: s})
	}
	return view
}

func statusWord(up bool) string {
	if up {
		return " — fine"
	}
	return " — no answer"
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
