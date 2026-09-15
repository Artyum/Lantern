package server

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lantern/internal/config"
	"lantern/internal/icons"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func testServerDir(t *testing.T) (http.Handler, string) {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	body := `{
		"title": "Test LAN",
		"sections": [{"name":"Infra","items":[{"name":"Traefik","url":"http://traefik.lan"}]}]
	}`
	if err := os.WriteFile(cfgPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	store, err := config.NewStore(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	cache := filepath.Join(dir, "cache")
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 404,
			Body:       io.NopCloser(strings.NewReader("no")),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})}
	res, err := icons.NewResolver(cache, client)
	if err != nil {
		t.Fatal(err)
	}
	iconDir := filepath.Join(cache, "icons")
	if err := os.WriteFile(filepath.Join(iconDir, "traefik.svg"), []byte("<svg id='t'/>"), 0o644); err != nil {
		t.Fatal(err)
	}
	return New(store, res), dir
}

func testServer(t *testing.T) http.Handler {
	h, _ := testServerDir(t)
	return h
}

func TestHealthz(t *testing.T) {
	h := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 || rec.Body.String() != "ok\n" {
		t.Fatalf("%d %q", rec.Code, rec.Body.String())
	}
}

func TestAPIConfig(t *testing.T) {
	h := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("code %d", rec.Code)
	}
	var out APIConfig
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Title != "Test LAN" || len(out.Sections) != 1 {
		t.Fatalf("%+v", out)
	}
	item := out.Sections[0].Items[0]
	if item.Name != "Traefik" || !strings.HasPrefix(item.Icon, "/icons/traefik") {
		t.Fatalf("%+v", item)
	}
}

func TestServeCachedIconAndFallback(t *testing.T) {
	h := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/icons/traefik", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "image/svg+xml" {
		t.Fatalf("%d %s %q", rec.Code, rec.Header().Get("Content-Type"), rec.Body.String())
	}
	if rec.Body.String() != "<svg id='t'/>" {
		t.Fatalf("body %q", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/icons/missing", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("missing icon %d %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "image/svg+xml" {
		t.Fatalf("missing icon content-type %q", rec.Header().Get("Content-Type"))
	}

	req = httptest.NewRequest(http.MethodGet, "/icons/__fallback", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "image/svg+xml" {
		t.Fatalf("fallback %d %s", rec.Code, rec.Header().Get("Content-Type"))
	}
}

func TestUpdateItemJSON(t *testing.T) {
	h, dir := testServerDir(t)
	body := `{"section":"Infra","originalUrl":"http://traefik.lan","name":"Traefik Proxy","url":"https://traefik.lan","icon":"traefik"}`
	req := httptest.NewRequest(http.MethodPut, "/api/item", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("code %d %s", rec.Code, rec.Body.String())
	}
	raw, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"name": "Traefik Proxy"`) || !strings.Contains(string(raw), `"url": "https://traefik.lan"`) {
		t.Fatalf("config %s", raw)
	}
}

func TestUpdateItemIconUpload(t *testing.T) {
	h, dir := testServerDir(t)
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("section", "Infra")
	_ = w.WriteField("originalUrl", "http://traefik.lan")
	_ = w.WriteField("name", "Traefik")
	_ = w.WriteField("url", "http://traefik.lan")
	fw, err := w.CreateFormFile("iconFile", "logo.svg")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write([]byte(`<svg id="up"/></svg>`)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPut, "/api/item", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("code %d %s", rec.Code, rec.Body.String())
	}
	var out APIItem
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out.IconSource, "upload-") || !strings.HasPrefix(out.Icon, "/icons/"+out.IconSource) {
		t.Fatalf("%+v", out)
	}
	iconPath := strings.Split(out.Icon, "?")[0]
	req = httptest.NewRequest(http.MethodGet, iconPath, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `id="up"`) {
		t.Fatalf("icon %d %q", rec.Code, rec.Body.String())
	}
	raw, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), out.IconSource) {
		t.Fatalf("config missing upload key: %s", raw)
	}
}

func TestAddAndDeleteSectionAPI(t *testing.T) {
	h, dir := testServerDir(t)
	req := httptest.NewRequest(http.MethodPost, "/api/section", strings.NewReader(`{"name":"Nowa"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("add %d %s", rec.Code, rec.Body.String())
	}
	raw, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"name": "Nowa"`) {
		t.Fatalf("config %s", raw)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/section", strings.NewReader(`{"name":"nowa"}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("dup %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/section", strings.NewReader(`{"name":"Infra"}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("del %d %s", rec.Code, rec.Body.String())
	}
	raw, err = os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "Traefik") || strings.Contains(string(raw), `"Infra"`) {
		t.Fatalf("still present: %s", raw)
	}
}

func TestSetLayoutAPI(t *testing.T) {
	h, dir := testServerDir(t)
	addReq := httptest.NewRequest(http.MethodPost, "/api/item", strings.NewReader(`{"section":"Infra","name":"Pi-hole","url":"http://dns.lan","icon":"pihole"}`))
	addReq.Header.Set("Content-Type", "application/json")
	addRec := httptest.NewRecorder()
	h.ServeHTTP(addRec, addReq)
	if addRec.Code != 200 {
		t.Fatalf("add item %d %s", addRec.Code, addRec.Body.String())
	}

	secReq := httptest.NewRequest(http.MethodPost, "/api/section", strings.NewReader(`{"name":"Media"}`))
	secReq.Header.Set("Content-Type", "application/json")
	secRec := httptest.NewRecorder()
	h.ServeHTTP(secRec, secReq)
	if secRec.Code != 200 {
		t.Fatalf("add section %d %s", secRec.Code, secRec.Body.String())
	}

	body := `{
		"sections": [
			{"name":"Media","items":["http://traefik.lan"]},
			{"name":"Infra","items":["http://dns.lan"]}
		]
	}`
	req := httptest.NewRequest(http.MethodPut, "/api/layout", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("layout %d %s", rec.Code, rec.Body.String())
	}
	raw, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	idxMedia := strings.Index(string(raw), `"name": "Media"`)
	idxTraefik := strings.Index(string(raw), `"url": "http://traefik.lan"`)
	idxInfra := strings.Index(string(raw), `"name": "Infra"`)
	if idxMedia == -1 || idxTraefik == -1 || idxInfra == -1 || idxTraefik < idxMedia || idxInfra < idxTraefik {
		t.Fatalf("expected traefik moved to Media before Infra: %s", raw)
	}
}

func TestSetLayoutAPIWithPositions(t *testing.T) {
	h, dir := testServerDir(t)
	addReq := httptest.NewRequest(http.MethodPost, "/api/item", strings.NewReader(`{"section":"Infra","name":"Pi-hole","url":"http://dns.lan","icon":"pihole"}`))
	addReq.Header.Set("Content-Type", "application/json")
	addRec := httptest.NewRecorder()
	h.ServeHTTP(addRec, addReq)
	if addRec.Code != 200 {
		t.Fatalf("add item %d %s", addRec.Code, addRec.Body.String())
	}

	body := `{
		"sections": [{
			"name": "Infra",
			"items": [
				{"url": "http://traefik.lan", "col": 2, "row": 1},
				{"url": "http://dns.lan", "col": 1, "row": 2}
			]
		}]
	}`
	req := httptest.NewRequest(http.MethodPut, "/api/layout", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("layout %d %s", rec.Code, rec.Body.String())
	}
	raw, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if !strings.Contains(text, `"col": 2`) || !strings.Contains(text, `"row": 2`) {
		t.Fatalf("expected stored positions: %s", text)
	}
}

func TestAddAndDeleteItemAPI(t *testing.T) {
	h, dir := testServerDir(t)
	req := httptest.NewRequest(http.MethodPost, "/api/item", strings.NewReader(`{"section":"Infra","name":"Pi-hole","url":"http://dns.lan","icon":"pihole"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("add %d %s", rec.Code, rec.Body.String())
	}
	raw, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"name": "Pi-hole"`) || !strings.Contains(string(raw), "http://dns.lan") {
		t.Fatalf("config %s", raw)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/item", strings.NewReader(`{"section":"Infra","name":"Dup","url":"http://traefik.lan"}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("dup %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/item", strings.NewReader(`{"section":"Infra","url":"http://traefik.lan"}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("del %d %s", rec.Code, rec.Body.String())
	}
	raw, err = os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "Traefik") || !strings.Contains(string(raw), "Pi-hole") {
		t.Fatalf("config %s", raw)
	}
}

func TestRefreshItemIconAPI(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	body := `{
		"title": "Test LAN",
		"sections": [{"name":"Infra","items":[{"name":"Traefik","url":"http://traefik.lan"}]}]
	}`
	if err := os.WriteFile(cfgPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	store, err := config.NewStore(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	cache := filepath.Join(dir, "cache")
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		code := 404
		out := "no"
		if strings.Contains(req.URL.Path, "/dashboard-icons/svg/traefik.svg") {
			code = 200
			out = `<svg id="fresh"/>`
		}
		return &http.Response{
			StatusCode: code,
			Body:       io.NopCloser(strings.NewReader(out)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})}
	res, err := icons.NewResolver(cache, client)
	if err != nil {
		t.Fatal(err)
	}
	iconDir := filepath.Join(cache, "icons")
	if err := os.WriteFile(filepath.Join(iconDir, "traefik.svg"), []byte("<svg id='stale'/>"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := New(store, res)

	req := httptest.NewRequest(http.MethodPost, "/api/item/icon", strings.NewReader(`{"name":"Traefik","url":"http://traefik.lan"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("code %d %s", rec.Code, rec.Body.String())
	}
	var out APIItem
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out.Icon, "/icons/traefik") {
		t.Fatalf("%+v", out)
	}
	iconPath := strings.Split(out.Icon, "?")[0]
	req = httptest.NewRequest(http.MethodGet, iconPath, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `id="fresh"`) {
		t.Fatalf("icon %d %q", rec.Code, rec.Body.String())
	}
}
