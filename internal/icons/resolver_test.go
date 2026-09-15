package icons

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lantern/internal/config"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestItemKey(t *testing.T) {
	if got := ItemKey(config.Item{Name: "Jellyfin", URL: "http://jellyfin.lan"}); got != "jellyfin" {
		t.Fatalf("got %q", got)
	}
	if got := ItemKey(config.Item{Name: "Uptime Kuma", URL: "http://kuma.lan"}); got != "uptimekuma" {
		t.Fatalf("got %q", got)
	}
	if got := ItemKey(config.Item{Name: "Pi-hole", URL: "http://dns.lan"}); got != "pihole" {
		t.Fatalf("got %q", got)
	}
}

func TestSimpleSlug(t *testing.T) {
	cases := map[string]string{
		"Uptime Kuma": "uptimekuma",
		"Pi-hole":     "pihole",
		"draw.io":     "drawdotio",
		"1&1":         "1and1",
		"Fresh RSS":   "freshrss",
		"Calibre Web": "calibreweb",
	}
	for in, want := range cases {
		if got := simpleSlug(in); got != want {
			t.Fatalf("%q: got %q want %q", in, got, want)
		}
	}
}

func TestDashSlug(t *testing.T) {
	cases := map[string]string{
		"Uptime Kuma": "uptime-kuma",
		"Pi-hole":     "pi-hole",
		"Calibre Web": "calibre-web",
		"Fresh RSS":   "fresh-rss",
		"Portainer":   "portainer",
	}
	for in, want := range cases {
		if got := dashSlug(in); got != want {
			t.Fatalf("%q: got %q want %q", in, got, want)
		}
	}
	slugs := dashboardSlugs("Uptime Kuma")
	if len(slugs) != 2 || slugs[0] != "uptime-kuma" || slugs[1] != "uptimekuma" {
		t.Fatalf("slugs %v", slugs)
	}
}

func TestCatalogIndexUsesOfficialAliases(t *testing.T) {
	idx := indexSimpleIcons([]simpleIcon{{
		Title:   "diagrams.net",
		Slug:    "diagramsdotnet",
		Aliases: &simpleAliases{Aka: []string{"draw.io"}},
	}})
	if got := idx[simpleSlug("draw.io")]; got != "diagramsdotnet" {
		t.Fatalf("got %q", got)
	}
	if got := idx[simpleSlug("diagrams.net")]; got != "diagramsdotnet" {
		t.Fatalf("title got %q", got)
	}
}

func TestResolveUsesCatalogAlias(t *testing.T) {
	dir := t.TempDir()
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case strings.Contains(req.URL.Path, "/data/simple-icons.json"):
			body := `[{"title":"diagrams.net","slug":"diagramsdotnet","aliases":{"aka":["draw.io"]}}]`
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: req}, nil
		case strings.HasSuffix(req.URL.Path, "/drawdotio.svg"):
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader("no")), Header: make(http.Header), Request: req}, nil
		case strings.HasSuffix(req.URL.Path, "/diagramsdotnet.svg"):
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`<svg id="catalog"/>`)), Header: make(http.Header), Request: req}, nil
		default:
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader("no")), Header: make(http.Header), Request: req}, nil
		}
	})}
	r, err := NewResolver(dir, client)
	if err != nil {
		t.Fatal(err)
	}
	item := config.Item{Name: "draw.io", URL: "http://draw.lan"}
	if err := r.Resolve(item); err != nil {
		t.Fatal(err)
	}
	data, _, ok := r.Read(ItemKey(item))
	if !ok || !strings.Contains(string(data), "catalog") {
		t.Fatalf("got %q ok=%v", data, ok)
	}
}

func TestResolvePriorityManualThenDashboardThenFaviconThenSimple(t *testing.T) {
	dir := t.TempDir()
	var hits []string
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		hits = append(hits, req.URL.String())
		body := "not-used"
		code := 404
		switch {
		case strings.Contains(req.URL.String(), "manual.example/icon.svg"):
			body = `<svg id="manual"/>`
			code = 200
		case strings.Contains(req.URL.Path, "/dashboard-icons/svg/grafana.svg"):
			body = `<svg id="dashboard"/>`
			code = 200
		case strings.Contains(req.URL.Path, "/dashboard-icons/svg/uptime-kuma.svg"):
			body = `<svg id="kuma"/>`
			code = 200
		case strings.Contains(req.URL.String(), "simple-icons"):
			body = `<svg id="simple"/>`
			code = 200
		case strings.HasSuffix(req.URL.Path, "/favicon.ico"):
			body = "ICO"
			code = 200
		}
		return &http.Response{
			StatusCode: code,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})}

	r, err := NewResolver(dir, client)
	if err != nil {
		t.Fatal(err)
	}

	item := config.Item{Name: "Jellyfin", URL: "http://jellyfin.lan", Icon: "https://manual.example/icon.svg"}
	if err := r.Resolve(item); err != nil {
		t.Fatal(err)
	}
	key := ItemKey(item)
	data, _, ok := r.Read(key)
	if !ok || !strings.Contains(string(data), "manual") {
		t.Fatalf("expected manual icon, got %q ok=%v hits=%v", data, ok, hits)
	}

	dir2 := t.TempDir()
	r2, _ := NewResolver(dir2, client)
	item2 := config.Item{Name: "Grafana", URL: "http://grafana.lan"}
	if err := r2.Resolve(item2); err != nil {
		t.Fatal(err)
	}
	data, _, ok = r2.Read(ItemKey(item2))
	if !ok || !strings.Contains(string(data), "dashboard") {
		t.Fatalf("expected dashboard icon, got %q hits=%v", data, hits)
	}

	dirKuma := t.TempDir()
	rKuma, _ := NewResolver(dirKuma, client)
	itemKuma := config.Item{Name: "Uptime Kuma", URL: "http://kuma.lan"}
	if err := rKuma.Resolve(itemKuma); err != nil {
		t.Fatal(err)
	}
	data, _, ok = rKuma.Read(ItemKey(itemKuma))
	if !ok || !strings.Contains(string(data), "kuma") {
		t.Fatalf("expected kebab dashboard icon, got %q", data)
	}

	failDash := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := "no"
		code := 404
		ct := "text/plain"
		if strings.Contains(req.URL.Path, "/dashboard-icons/") || strings.Contains(req.URL.String(), "simple-icons") {
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader("no")), Header: make(http.Header), Request: req}, nil
		}
		if strings.HasSuffix(req.URL.Path, "/favicon.ico") {
			body = "FAV"
			code = 200
			ct = "image/x-icon"
		}
		if strings.Contains(req.URL.Host, "onlyfav.lan") && req.URL.Path == "/" {
			body = `<html><link rel="icon" href="/favicon.ico"></html>`
			code = 200
			ct = "text/html"
		}
		h := make(http.Header)
		h.Set("Content-Type", ct)
		return &http.Response{StatusCode: code, Body: io.NopCloser(strings.NewReader(body)), Header: h, Request: req}, nil
	})}
	dir3 := t.TempDir()
	r3, _ := NewResolver(dir3, failDash)
	item3 := config.Item{Name: "Unknownxyz", URL: "http://onlyfav.lan/"}
	if err := r3.Resolve(item3); err != nil {
		t.Fatal(err)
	}
	data, _, ok = r3.Read(ItemKey(item3))
	if !ok || string(data) != "FAV" {
		t.Fatalf("expected favicon, got %q ok=%v", data, ok)
	}

	onlySimple := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := "no"
		code := 404
		if strings.Contains(req.URL.String(), "simple-icons") && strings.HasSuffix(req.URL.Path, "/nosuchdash.svg") {
			body = `<svg id="simple"/>`
			code = 200
		}
		return &http.Response{StatusCode: code, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: req}, nil
	})}
	dir4 := t.TempDir()
	r4, _ := NewResolver(dir4, onlySimple)
	item4 := config.Item{Name: "No Such Dash", URL: "http://nosuch.lan"}
	if err := r4.Resolve(item4); err != nil {
		t.Fatal(err)
	}
	data, _, ok = r4.Read(ItemKey(item4))
	if !ok || !strings.Contains(string(data), "simple") {
		t.Fatalf("expected simple fallback, got %q", data)
	}
}

func TestRefreshReplacesCachedIcon(t *testing.T) {
	dir := t.TempDir()
	n := 0
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		n++
		body := `<svg id="first"/>`
		if n >= 2 {
			body = `<svg id="second"/>`
		}
		code := 404
		if strings.Contains(req.URL.Path, "/dashboard-icons/svg/grafana.svg") {
			code = 200
		}
		return &http.Response{
			StatusCode: code,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})}
	r, err := NewResolver(dir, client)
	if err != nil {
		t.Fatal(err)
	}
	item := config.Item{Name: "Grafana", URL: "http://grafana.lan"}
	if err := r.Resolve(item); err != nil {
		t.Fatal(err)
	}
	if data, _, _ := r.Read(ItemKey(item)); !strings.Contains(string(data), "first") {
		t.Fatalf("got %q", data)
	}
	if err := r.Refresh(item); err != nil {
		t.Fatal(err)
	}
	if data, _, _ := r.Read(ItemKey(item)); !strings.Contains(string(data), "second") {
		t.Fatalf("got %q", data)
	}
}

func TestCacheVersionClearsStaleIcons(t *testing.T) {
	dir := t.TempDir()
	iconsDir := filepath.Join(dir, "icons")
	if err := os.MkdirAll(iconsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(iconsDir, "portainer.svg"), []byte("<svg id='old'/>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(iconsDir, "upload-abc.svg"), []byte("<svg id='keep'/>"), 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := NewResolver(dir, http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, ok := r.Read("portainer"); ok {
		t.Fatal("expected stale dashboard cache to be cleared")
	}
	data, _, ok := r.Read("upload-abc")
	if !ok || !strings.Contains(string(data), "keep") {
		t.Fatalf("expected upload to remain, got %q ok=%v", data, ok)
	}
}

func TestBlockedFileScheme(t *testing.T) {
	dir := t.TempDir()
	r, _ := NewResolver(dir, http.DefaultClient)
	_, _, err := r.fetchBytes("file:///etc/passwd")
	if err == nil {
		t.Fatal("expected blocked file url")
	}
}

func TestNegativeCacheNoRetryStorm(t *testing.T) {
	dir := t.TempDir()
	n := 0
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		n++
		return &http.Response{
			StatusCode: 404,
			Body:       io.NopCloser(strings.NewReader("no")),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})}
	r, _ := NewResolver(dir, client)
	item := config.Item{Name: "Nope", URL: "http://nope.lan"}
	_ = r.Resolve(item)
	first := n
	_ = r.Resolve(item)
	if n != first {
		t.Fatalf("retried after failure: %d -> %d", first, n)
	}
}

func TestServeCachedFileViaRead(t *testing.T) {
	dir := t.TempDir()
	r, _ := NewResolver(dir, http.DefaultClient)
	p := filepath.Join(dir, "icons", "jellyfin.svg")
	if err := os.WriteFile(p, []byte("<svg/>"), 0o644); err != nil {
		t.Fatal(err)
	}
	data, ct, ok := r.Read("jellyfin")
	if !ok || string(data) != "<svg/>" || ct != "image/svg+xml" {
		t.Fatalf("got %q %q %v", data, ct, ok)
	}
	if r.Revision("jellyfin") == "" {
		t.Fatal("expected revision for cached icon")
	}
	if r.Revision("missing") != "" {
		t.Fatal("expected empty revision for missing icon")
	}
}

func TestFaviconFromHTML(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "text/html")
			_, _ = io.WriteString(w, `<link rel="shortcut icon" href="/brand.png">`)
		case "/brand.png":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write([]byte("PNGDATA"))
		default:
			http.NotFound(w, req)
		}
	}))
	defer ts.Close()

	dir := t.TempDir()
	r, _ := NewResolver(dir, ts.Client())
	item := config.Item{Name: "Customsvc", URL: ts.URL + "/"}
	// force skip simple icons by making client still hit jsdelivr 404 via default client on ts only
	// Use a mux client that 404s simple-icons.
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Host, "jsdelivr") {
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header), Request: req}, nil
		}
		rt := ts.Client().Transport
		if rt == nil {
			rt = http.DefaultTransport
		}
		return rt.RoundTrip(req)
	})}
	r.client = client
	if err := r.Resolve(item); err != nil {
		t.Fatal(err)
	}
	data, _, ok := r.Read(ItemKey(item))
	if !ok || string(data) != "PNGDATA" {
		t.Fatalf("got %q ok=%v", data, ok)
	}
}
