// Package quadlet reads the few fields of a production .container unit that
// the status and ports tooling needs. It is not a general systemd parser.
package quadlet

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Port is one published port. Host is the port on the box.
type Port struct {
	IP    string // bind address, "" when published on all interfaces
	Host  int
	Proto string // tcp or udp
}

// External reports whether the port can be reached from another machine.
func (p Port) External() bool { return p.IP != "127.0.0.1" && p.IP != "::1" }

// Unit is one parsed .container file.
type Unit struct {
	Name        string // file name without .container, e.g. openpak-wonder
	Short       string // Name without the openpak- prefix
	Description string
	Image       string
	HostNetwork bool
	Published   []Port
	Marker      []int // "# openpak-ports=A B C" (auth, secure, health)
	HealthPort  int   // 0 when unknown
}

var (
	markerRe = regexp.MustCompile(`^#\s*openpak-ports=([0-9 ]+)\s*$`)
	healthRe = regexp.MustCompile(`HEALTH_LISTEN=[^ ]*:(\d+)`)
	descRe   = regexp.MustCompile(`(?i)(?:health|HTTP)\s+(\d{4,5})`)
)

// Load reads every *.container file in dir, sorted by name.
func Load(dir string) ([]Unit, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.container"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	var out []Unit
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		out = append(out, Parse(strings.TrimSuffix(filepath.Base(f), ".container"), string(b)))
	}
	return out, nil
}

// Parse reads one unit.
func Parse(name, text string) Unit {
	u := Unit{Name: name, Short: strings.TrimPrefix(name, "openpak-")}
	var env []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if m := markerRe.FindStringSubmatch(line); m != nil {
			for _, f := range strings.Fields(m[1]) {
				n, _ := strconv.Atoi(f)
				u.Marker = append(u.Marker, n)
			}
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok || strings.HasPrefix(line, "#") {
			continue
		}
		switch k {
		case "Description":
			u.Description = v
		case "Image":
			u.Image = v
		case "Network":
			if v == "host" {
				u.HostNetwork = true
			}
		case "Environment":
			env = append(env, v)
		case "PublishPort":
			if p, ok := parsePort(v); ok {
				u.Published = append(u.Published, p)
			}
		}
	}
	if m := healthRe.FindStringSubmatch(strings.Join(env, " ")); m != nil {
		u.HealthPort, _ = strconv.Atoi(m[1])
	} else if m := descRe.FindStringSubmatch(u.Description); m != nil {
		u.HealthPort, _ = strconv.Atoi(m[1])
	} else if len(u.Marker) >= 3 {
		u.HealthPort = u.Marker[2]
	}
	return u
}

func parsePort(spec string) (Port, bool) {
	p := Port{Proto: "tcp"}
	if a, b, ok := strings.Cut(spec, "/"); ok {
		spec, p.Proto = a, b
	}
	f := strings.Split(spec, ":")
	if len(f) < 2 {
		return p, false
	}
	n, err := strconv.Atoi(f[len(f)-2])
	if err != nil {
		return p, false
	}
	p.Host = n
	if len(f) == 3 && f[0] != "0.0.0.0" {
		p.IP = f[0]
	}
	return p, true
}

// Image returns the image without registry, e.g. "nex:latest".
func (u Unit) ImageShort() string {
	i := strings.TrimPrefix(u.Image, "ghcr.io/openpak/")
	i = strings.TrimPrefix(i, "docker.io/library/")
	if j := strings.Index(i, "@sha256:"); j >= 0 {
		i = i[:j] + "@sha256"
	}
	return i
}
