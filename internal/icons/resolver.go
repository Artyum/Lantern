package icons

import (
	"crypto/tls"
	"embed"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"lantern/internal/config"
)

//go:embed logo.svg
var logoFS embed.FS

const (
	maxBody            = 512 * 1024
	maxRedirects       = 3
	cacheSourceVersion = "2"
	simpleIconsURL     = "https://cdn.jsdelivr.net/npm/simple-icons/icons/%s.svg"
	dashboardIconsSVG  = "https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons/svg/%s.svg"
	dashboardIconsPNG  = "https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons/png/%s.png"
)

var (
	linkTagRe = regexp.MustCompile(`(?is)<link\b[^>]*>`)
	relRe     = regexp.MustCompile(`(?i)\brel\s*=\s*["']([^"']+)["']`)
	hrefRe    = regexp.MustCompile(`(?i)\bhref\s*=\s*["']([^"']+)["']`)
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Resolver struct {
	dir    string
	client HTTPClient

	mu          sync.Mutex
	failed      map[string]struct{}
	catalogOnce sync.Once
	catalog     map[string]string
}

func LogoSVG() []byte {
	b, err := logoFS.ReadFile("logo.svg")
	if err != nil {
		return []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"/>`)
	}
	return b
}

func NewResolver(dir string, client HTTPClient) (*Resolver, error) {
	if client == nil {
		client = &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: true}, // homelab self-signed
				Proxy:               http.ProxyFromEnvironment,
				TLSHandshakeTimeout: 5 * time.Second,
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= maxRedirects {
					return fmt.Errorf("too many redirects")
				}
				if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
					return fmt.Errorf("blocked redirect scheme")
				}
				return nil
			},
		}
	}
	r := &Resolver{dir: dir, client: client, failed: map[string]struct{}{}}
	if dir != "" {
		if err := os.MkdirAll(r.iconsDir(), 0o755); err != nil {
			return nil, err
		}
		if err := r.migrateCache(); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r *Resolver) migrateCache() error {
	if r.dir == "" {
		return nil
	}
	iconsDir := r.iconsDir()
	marker := filepath.Join(iconsDir, ".lantern-icon-cache")
	if b, err := os.ReadFile(marker); err == nil && strings.TrimSpace(string(b)) == cacheSourceVersion {
		return nil
	}
	entries, err := os.ReadDir(iconsDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "upload-") {
			continue
		}
		_ = os.RemoveAll(filepath.Join(iconsDir, name))
	}
	return os.WriteFile(marker, []byte(cacheSourceVersion+"\n"), 0o644)
}

func (r *Resolver) iconsDir() string {
	return filepath.Join(r.dir, "icons")
}

func ItemKey(item config.Item) string {
	if isHTTPURL(item.Icon) {
		return slugify(hostOf(item.Icon) + "-" + path.Base(item.Icon))
	}
	if item.Icon != "" {
		return slugify(item.Icon)
	}
	if s := simpleSlug(item.Name); s != "" {
		return s
	}
	return slugify(hostOf(item.URL))
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			prevDash = false
		case r == '.' || r == '_' || r == '-':
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		default:
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "icon"
	}
	return out
}

func isHTTPURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func hostOf(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return u.Hostname()
}

func (r *Resolver) CachePath(key string) string {
	if r.dir == "" {
		return ""
	}
	return filepath.Join(r.dir, "icons", key)
}

func (r *Resolver) Read(key string) (data []byte, contentType string, ok bool) {
	p := r.CachePath(key)
	if p == "" {
		return nil, "", false
	}
	matches, _ := filepath.Glob(p + ".*")
	if len(matches) == 0 {
		if _, err := os.Stat(p); err == nil {
			matches = []string{p}
		}
	}
	if len(matches) == 0 {
		return nil, "", false
	}
	b, err := os.ReadFile(matches[0])
	if err != nil {
		return nil, "", false
	}
	return b, typeFromName(matches[0], b), true
}

func (r *Resolver) Revision(key string) string {
	p := r.CachePath(key)
	if p == "" {
		return ""
	}
	matches, _ := filepath.Glob(p + ".*")
	if len(matches) == 0 {
		if _, err := os.Stat(p); err == nil {
			matches = []string{p}
		}
	}
	if len(matches) == 0 {
		return ""
	}
	info, err := os.Stat(matches[0])
	if err != nil {
		return ""
	}
	return strconv.FormatInt(info.ModTime().Unix(), 10)
}

func typeFromName(name string, body []byte) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".ico":
		return "image/x-icon"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	}
	ct := http.DetectContentType(body)
	if strings.Contains(string(body[:min(len(body), 200)]), "<svg") {
		return "image/svg+xml"
	}
	return ct
}

func (r *Resolver) Warmup(cfg *config.Config) {
	if cfg == nil {
		return
	}
	for _, sec := range cfg.Sections {
		for _, item := range sec.Items {
			item := item
			_ = r.Resolve(item)
		}
	}
}

func (r *Resolver) Resolve(item config.Item) error {
	key := ItemKey(item)
	if _, _, ok := r.Read(key); ok {
		return nil
	}
	if r.rememberedFail(key) {
		return fmt.Errorf("icon %s previously failed", key)
	}

	if isHTTPURL(item.Icon) {
		body, ext, err := r.fetchBytes(item.Icon)
		if err == nil && len(body) > 0 {
			return r.write(key, ext, body)
		}
		r.markFail(key)
		return err
	}

	names := iconNames(item)
	for _, name := range names {
		body, ext, err := r.fetchDashboard(name)
		if err == nil && len(body) > 0 {
			return r.write(key, ext, body)
		}
	}
	if body, ext, err := r.fetchFavicon(item.URL); err == nil && len(body) > 0 {
		return r.write(key, ext, body)
	}
	for _, name := range names {
		body, ext, err := r.fetchSimpleFor(name)
		if err == nil && len(body) > 0 {
			return r.write(key, ext, body)
		}
	}

	r.markFail(key)
	return fmt.Errorf("no icon found for %q", item.Name)
}

func iconNames(item config.Item) []string {
	names := make([]string, 0, 2)
	if iconSlug(item.Icon) {
		names = append(names, item.Icon)
	}
	if item.Name != "" && (len(names) == 0 || names[0] != item.Name) {
		names = append(names, item.Name)
	}
	return names
}

func iconSlug(s string) bool {
	return s != "" && !isHTTPURL(s) && !strings.HasPrefix(s, "upload-")
}

func (r *Resolver) fetchDashboard(name string) ([]byte, string, error) {
	var last error
	for _, slug := range dashboardSlugs(name) {
		body, ext, err := r.fetchDashboardSlug(slug)
		if err == nil {
			return body, ext, nil
		}
		last = err
	}
	if last == nil {
		return nil, "", fmt.Errorf("dashboard icon not found")
	}
	return nil, "", last
}

func (r *Resolver) fetchDashboardSlug(slug string) ([]byte, string, error) {
	if !validDashboardSlug(slug) {
		return nil, "", fmt.Errorf("invalid dashboard icon slug")
	}
	body, ext, err := r.fetchBytes(fmt.Sprintf(dashboardIconsSVG, slug))
	if err == nil {
		return body, ext, nil
	}
	return r.fetchBytes(fmt.Sprintf(dashboardIconsPNG, slug))
}

func validDashboardSlug(slug string) bool {
	if slug == "" || len(slug) > 80 {
		return false
	}
	for _, r := range slug {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' {
			continue
		}
		return false
	}
	return true
}

func (r *Resolver) rememberedFail(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.failed[key]
	return ok
}

func (r *Resolver) markFail(key string) {
	r.mu.Lock()
	r.failed[key] = struct{}{}
	r.mu.Unlock()
}

func (r *Resolver) ForgetFail(key string) {
	r.mu.Lock()
	delete(r.failed, key)
	r.mu.Unlock()
}

func (r *Resolver) Store(key, ext string, body []byte) error {
	if key == "" || len(body) == 0 {
		return fmt.Errorf("empty icon")
	}
	r.ForgetFail(key)
	return r.write(key, ext, body)
}

func (r *Resolver) Drop(key string) {
	if key == "" {
		return
	}
	r.ForgetFail(key)
	p := r.CachePath(key)
	if p == "" {
		return
	}
	if matches, _ := filepath.Glob(p + ".*"); len(matches) > 0 {
		for _, old := range matches {
			_ = os.Remove(old)
		}
	}
	_ = os.Remove(p)
}

func (r *Resolver) Refresh(item config.Item) error {
	r.Drop(ItemKey(item))
	return r.Resolve(item)
}

func (r *Resolver) write(key, ext string, body []byte) error {
	p := r.CachePath(key)
	if p == "" {
		return nil
	}
	if ext == "" {
		ext = ".bin"
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	if matches, _ := filepath.Glob(p + ".*"); len(matches) > 0 {
		for _, old := range matches {
			_ = os.Remove(old)
		}
	}
	return os.WriteFile(p+ext, body, 0o644)
}

func (r *Resolver) fetchSimple(slug string) ([]byte, string, error) {
	if slug == "" {
		return nil, "", fmt.Errorf("empty slug")
	}
	return r.fetchBytes(fmt.Sprintf(simpleIconsURL, slug))
}

func (r *Resolver) fetchFavicon(pageURL string) ([]byte, string, error) {
	if !isHTTPURL(pageURL) {
		return nil, "", fmt.Errorf("invalid page url")
	}
	html, _, err := r.fetchBytes(pageURL)
	if err == nil {
		if href := firstIconHref(string(html)); href != "" {
			abs, err := resolveURL(pageURL, href)
			if err == nil {
				if body, ext, err := r.fetchBytes(abs); err == nil {
					return body, ext, nil
				}
			}
		}
	}
	base, err := url.Parse(pageURL)
	if err != nil {
		return nil, "", err
	}
	base.Path = "/favicon.ico"
	base.RawQuery = ""
	base.Fragment = ""
	return r.fetchBytes(base.String())
}

func firstIconHref(html string) string {
	for _, tag := range linkTagRe.FindAllString(html, -1) {
		rel := ""
		if m := relRe.FindStringSubmatch(tag); len(m) == 2 {
			rel = strings.ToLower(m[1])
		}
		if !strings.Contains(rel, "icon") {
			continue
		}
		if m := hrefRe.FindStringSubmatch(tag); len(m) == 2 {
			return strings.TrimSpace(m[1])
		}
	}
	return ""
}

func resolveURL(baseRaw, ref string) (string, error) {
	base, err := url.Parse(baseRaw)
	if err != nil {
		return "", err
	}
	u, err := url.Parse(ref)
	if err != nil {
		return "", err
	}
	if u.Scheme == "file" || (u.Scheme != "" && u.Scheme != "http" && u.Scheme != "https") {
		return "", fmt.Errorf("blocked scheme")
	}
	return base.ResolveReference(u).String(), nil
}

func (r *Resolver) fetchBytes(raw string) ([]byte, string, error) {
	return r.fetchBytesN(raw, maxBody)
}

func (r *Resolver) fetchBytesN(raw string, limit int) ([]byte, string, error) {
	if !isHTTPURL(raw) {
		return nil, "", fmt.Errorf("blocked url")
	}
	if limit <= 0 {
		limit = maxBody
	}
	req, err := http.NewRequest(http.MethodGet, raw, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "Lantern/1.0")
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, int64(limit)+1))
	if err != nil {
		return nil, "", err
	}
	if len(body) > limit {
		return nil, "", fmt.Errorf("body too large")
	}
	if len(body) == 0 {
		return nil, "", fmt.Errorf("empty body")
	}
	ext := extFromURL(raw, resp.Header.Get("Content-Type"), body)
	return body, ext, nil
}

func extFromURL(raw, contentType string, body []byte) string {
	ct := strings.ToLower(contentType)
	switch {
	case strings.Contains(ct, "svg"):
		return ".svg"
	case strings.Contains(ct, "png"):
		return ".png"
	case strings.Contains(ct, "jpeg") || strings.Contains(ct, "jpg"):
		return ".jpg"
	case strings.Contains(ct, "webp"):
		return ".webp"
	case strings.Contains(ct, "gif"):
		return ".gif"
	case strings.Contains(ct, "icon"):
		return ".ico"
	}
	p := strings.ToLower(path.Ext(raw))
	switch p {
	case ".svg", ".png", ".ico", ".jpg", ".jpeg", ".webp", ".gif":
		return p
	}
	if strings.Contains(string(body[:min(len(body), 256)]), "<svg") {
		return ".svg"
	}
	return ".bin"
}
