package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"lantern/internal/config"
	"lantern/internal/icons"
	"lantern/web"
)

const maxIconUpload = 512 * 1024

type APIItem struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	Icon       string `json:"icon"`
	IconSource string `json:"iconSource,omitempty"`
	Col        int    `json:"col,omitempty"`
	Row        int    `json:"row,omitempty"`
}

type APISection struct {
	Name  string    `json:"name"`
	Items []APIItem `json:"items"`
}

type APIConfig struct {
	Title    string       `json:"title"`
	Sections []APISection `json:"sections"`
}

type sectionNameBody struct {
	Name string `json:"name"`
}

type layoutItemBody struct {
	URL string `json:"url"`
	Col int    `json:"col"`
	Row int    `json:"row"`
}

func (l *layoutItemBody) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("empty layout item")
	}
	if data[0] == '"' {
		var url string
		if err := json.Unmarshal(data, &url); err != nil {
			return err
		}
		l.URL = url
		return nil
	}
	var raw struct {
		URL string `json:"url"`
		Col int    `json:"col"`
		Row int    `json:"row"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	l.URL = raw.URL
	l.Col = raw.Col
	l.Row = raw.Row
	return nil
}

type layoutSectionBody struct {
	Name  string           `json:"name"`
	Items []layoutItemBody `json:"items"`
}

type layoutBody struct {
	Sections []layoutSectionBody `json:"sections"`
}

type itemUpdate struct {
	Section     string `json:"section"`
	OriginalURL string `json:"originalUrl"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Icon        string `json:"icon"`
}

type Server struct {
	store *config.Store
	icons *icons.Resolver
	mux   *http.ServeMux
}

func New(store *config.Store, resolver *icons.Resolver) http.Handler {
	s := &Server{store: store, icons: resolver, mux: http.NewServeMux()}
	s.mux.HandleFunc("GET /healthz", s.healthz)
	s.mux.HandleFunc("GET /api/config", s.apiConfig)
	s.mux.HandleFunc("PUT /api/item", s.updateItem)
	s.mux.HandleFunc("POST /api/item", s.addItem)
	s.mux.HandleFunc("DELETE /api/item", s.deleteItem)
	s.mux.HandleFunc("POST /api/item/icon", s.refreshItemIcon)
	s.mux.HandleFunc("POST /api/section", s.addSection)
	s.mux.HandleFunc("DELETE /api/section", s.deleteSection)
	s.mux.HandleFunc("PUT /api/layout", s.setLayout)
	s.mux.HandleFunc("GET /icons/{key}", s.icon)
	static, err := fs.Sub(web.FS, ".")
	if err != nil {
		static = web.FS
	}
	fileServer := http.FileServer(http.FS(static))
	s.mux.Handle("GET /", s.staticAssetsHandler(fileServer))
	return s.mux
}

func (s *Server) staticAssetsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			s.index(w, r)
			return
		}
		switch r.URL.Path {
		case "/app.css", "/app.js":
			w.Header().Set("Cache-Control", "no-cache")
		default:
			if strings.HasPrefix(r.URL.Path, "/fonts/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok\n"))
}

func (s *Server) apiConfig(w http.ResponseWriter, _ *http.Request) {
	out, err := s.currentAPIConfig()
	if err != nil {
		http.Error(w, "config unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(out)
}

func (s *Server) index(w http.ResponseWriter, _ *http.Request) {
	out, err := s.currentAPIConfig()
	if err != nil {
		http.Error(w, "config unavailable", http.StatusInternalServerError)
		return
	}
	configJSON, err := json.Marshal(out)
	if err != nil {
		http.Error(w, "config unavailable", http.StatusInternalServerError)
		return
	}
	page, err := web.FS.ReadFile("index.html")
	if err != nil {
		http.Error(w, "page unavailable", http.StatusInternalServerError)
		return
	}
	page = []byte(strings.Replace(
		string(page),
		"<!-- initial-config -->",
		`<script id="initial-config" type="application/json">`+string(configJSON)+`</script>`,
		1,
	))
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(page)
}

func (s *Server) currentAPIConfig() (APIConfig, error) {
	cfg, reloaded, err := s.store.Get()
	if err != nil {
		return APIConfig{}, err
	}
	if reloaded {
		go s.icons.Warmup(cfg)
	}
	out := APIConfig{Title: cfg.Title, Sections: make([]APISection, 0, len(cfg.Sections))}
	for _, sec := range cfg.Sections {
		as := APISection{Name: sec.Name, Items: make([]APIItem, 0, len(sec.Items))}
		for _, item := range sec.Items {
			as.Items = append(as.Items, s.toAPIItem(item))
		}
		out.Sections = append(out.Sections, as)
	}
	return out, nil
}

func (s *Server) toAPIItem(item config.Item) APIItem {
	key := icons.ItemKey(item)
	icon := "/icons/" + key
	if rev := s.icons.Revision(key); rev != "" {
		icon += "?v=" + rev
	}
	return APIItem{
		Name:       item.Name,
		URL:        item.URL,
		Icon:       icon,
		IconSource: item.Icon,
		Col:        item.Col,
		Row:        item.Row,
	}
}

func (s *Server) updateItem(w http.ResponseWriter, r *http.Request) {
	in, fileName, fileBody, err := parseItemUpdate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	in.Section = strings.TrimSpace(in.Section)
	in.OriginalURL = strings.TrimSpace(in.OriginalURL)
	in.Name = strings.TrimSpace(in.Name)
	in.URL = strings.TrimSpace(in.URL)
	in.Icon = strings.TrimSpace(in.Icon)
	if in.Section == "" || in.OriginalURL == "" {
		http.Error(w, "section and originalUrl are required", http.StatusBadRequest)
		return
	}

	existing, ok := s.findItem(in.Section, in.OriginalURL)
	if !ok {
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}

	next := config.Item{
		Name: in.Name,
		URL:  in.URL,
		Icon: in.Icon,
		Col:  existing.Col,
		Row:  existing.Row,
	}
	if len(fileBody) > 0 {
		key := existing.Icon
		if !strings.HasPrefix(key, "upload-") || !validKey(key) {
			key = newUploadKey()
		}
		ext, err := uploadExt(fileName, fileBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.icons.Store(key, ext, fileBody); err != nil {
			http.Error(w, "could not store icon", http.StatusInternalServerError)
			return
		}
		next.Icon = key
	}

	if err := s.store.UpdateItem(in.Section, in.OriginalURL, next); err != nil {
		http.Error(w, err.Error(), itemErrorStatus(err))
		return
	}

	s.icons.ForgetFail(icons.ItemKey(next))
	go func(item config.Item) { _ = s.icons.Resolve(item) }(next)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(s.toAPIItem(next))
}

func (s *Server) addItem(w http.ResponseWriter, r *http.Request) {
	in, fileName, fileBody, err := parseItemUpdate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	in.Section = strings.TrimSpace(in.Section)
	in.Name = strings.TrimSpace(in.Name)
	in.URL = strings.TrimSpace(in.URL)
	in.Icon = strings.TrimSpace(in.Icon)
	if in.Section == "" {
		http.Error(w, "section is required", http.StatusBadRequest)
		return
	}

	next := config.Item{
		Name: in.Name,
		URL:  in.URL,
		Icon: in.Icon,
	}
	if len(fileBody) > 0 {
		key := newUploadKey()
		ext, err := uploadExt(fileName, fileBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.icons.Store(key, ext, fileBody); err != nil {
			http.Error(w, "could not store icon", http.StatusInternalServerError)
			return
		}
		next.Icon = key
	}

	if err := s.store.AddItem(in.Section, next); err != nil {
		http.Error(w, err.Error(), itemErrorStatus(err))
		return
	}

	s.icons.ForgetFail(icons.ItemKey(next))
	go func(item config.Item) { _ = s.icons.Resolve(item) }(next)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(s.toAPIItem(next))
}

func (s *Server) refreshItemIcon(w http.ResponseWriter, r *http.Request) {
	var in itemUpdate
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	next := config.Item{
		Name: strings.TrimSpace(in.Name),
		URL:  strings.TrimSpace(in.URL),
	}
	if next.Name == "" || next.URL == "" {
		http.Error(w, "name and url are required", http.StatusBadRequest)
		return
	}
	if err := s.icons.Refresh(next); err != nil {
		http.Error(w, "could not refresh icon", http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(s.toAPIItem(next))
}

func (s *Server) deleteItem(w http.ResponseWriter, r *http.Request) {
	var in itemUpdate
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	section := strings.TrimSpace(in.Section)
	itemURL := strings.TrimSpace(in.URL)
	if itemURL == "" {
		itemURL = strings.TrimSpace(in.OriginalURL)
	}
	if section == "" || itemURL == "" {
		http.Error(w, "section and url are required", http.StatusBadRequest)
		return
	}
	if err := s.store.DeleteItem(section, itemURL); err != nil {
		http.Error(w, err.Error(), itemErrorStatus(err))
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func itemErrorStatus(err error) int {
	switch {
	case errors.Is(err, config.ErrItemNotFound), errors.Is(err, config.ErrSectionNotFound):
		return http.StatusNotFound
	case errors.Is(err, config.ErrItemExists):
		return http.StatusConflict
	default:
		return http.StatusBadRequest
	}
}

func layoutErrorStatus(err error) int {
	switch {
	case errors.Is(err, config.ErrItemNotFound), errors.Is(err, config.ErrSectionNotFound):
		return http.StatusNotFound
	default:
		return http.StatusBadRequest
	}
}

func (s *Server) addSection(w http.ResponseWriter, r *http.Request) {
	var in sectionNameBody
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if len(name) > 80 {
		http.Error(w, "name too long", http.StatusBadRequest)
		return
	}
	if err := s.store.AddSection(name); err != nil {
		code := http.StatusBadRequest
		if errors.Is(err, config.ErrSectionExists) {
			code = http.StatusConflict
		}
		http.Error(w, err.Error(), code)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(APISection{Name: name, Items: []APIItem{}})
}

func (s *Server) setLayout(w http.ResponseWriter, r *http.Request) {
	var in layoutBody
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	layout := make([]config.LayoutSection, 0, len(in.Sections))
	for _, sec := range in.Sections {
		items := make([]config.LayoutItem, 0, len(sec.Items))
		for _, item := range sec.Items {
			items = append(items, config.LayoutItem{
				URL: strings.TrimSpace(item.URL),
				Col: item.Col,
				Row: item.Row,
			})
		}
		layout = append(layout, config.LayoutSection{
			Name:  strings.TrimSpace(sec.Name),
			Items: items,
		})
	}
	if err := s.store.SetLayout(layout); err != nil {
		http.Error(w, err.Error(), layoutErrorStatus(err))
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func (s *Server) deleteSection(w http.ResponseWriter, r *http.Request) {
	var in sectionNameBody
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if err := s.store.DeleteSection(name); err != nil {
		code := http.StatusBadRequest
		if errors.Is(err, config.ErrSectionNotFound) {
			code = http.StatusNotFound
		}
		http.Error(w, err.Error(), code)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func (s *Server) findItem(sectionName, itemURL string) (config.Item, bool) {
	cfg, _, err := s.store.Get()
	if err != nil {
		return config.Item{}, false
	}
	for _, sec := range cfg.Sections {
		if sec.Name != sectionName {
			continue
		}
		for _, item := range sec.Items {
			if item.URL == itemURL {
				return item, true
			}
		}
	}
	return config.Item{}, false
}

func parseItemUpdate(r *http.Request) (itemUpdate, string, []byte, error) {
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		if err := r.ParseMultipartForm(maxIconUpload + 64*1024); err != nil {
			return itemUpdate{}, "", nil, fmt.Errorf("invalid form")
		}
		in := itemUpdate{
			Section:     r.FormValue("section"),
			OriginalURL: r.FormValue("originalUrl"),
			Name:        r.FormValue("name"),
			URL:         r.FormValue("url"),
			Icon:        r.FormValue("icon"),
		}
		file, hdr, err := r.FormFile("iconFile")
		if err != nil {
			if errors.Is(err, http.ErrMissingFile) {
				return in, "", nil, nil
			}
			return itemUpdate{}, "", nil, fmt.Errorf("invalid icon file")
		}
		defer file.Close()
		body, err := io.ReadAll(io.LimitReader(file, maxIconUpload+1))
		if err != nil {
			return itemUpdate{}, "", nil, fmt.Errorf("could not read icon")
		}
		if len(body) == 0 {
			return itemUpdate{}, "", nil, fmt.Errorf("empty icon file")
		}
		if len(body) > maxIconUpload {
			return itemUpdate{}, "", nil, fmt.Errorf("icon too large")
		}
		name := ""
		if hdr != nil {
			name = hdr.Filename
		}
		return in, name, body, nil
	}

	var in itemUpdate
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&in); err != nil {
		return itemUpdate{}, "", nil, fmt.Errorf("invalid json")
	}
	return in, "", nil, nil
}

func newUploadKey() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "upload-icon"
	}
	return "upload-" + hex.EncodeToString(b[:])
}

func uploadExt(filename string, body []byte) (string, error) {
	ext := strings.ToLower(path.Ext(filename))
	switch ext {
	case ".svg", ".png", ".ico", ".jpg", ".webp", ".gif":
		return ext, nil
	case ".jpeg":
		return ".jpg", nil
	}
	head := body
	if len(head) > 256 {
		head = head[:256]
	}
	if strings.Contains(strings.ToLower(string(head)), "<svg") {
		return ".svg", nil
	}
	ct := http.DetectContentType(body)
	switch {
	case strings.Contains(ct, "png"):
		return ".png", nil
	case strings.Contains(ct, "jpeg"):
		return ".jpg", nil
	case strings.Contains(ct, "gif"):
		return ".gif", nil
	case strings.Contains(ct, "webp"):
		return ".webp", nil
	case strings.Contains(ct, "icon"):
		return ".ico", nil
	}
	return "", fmt.Errorf("unsupported icon type")
}

func (s *Server) itemByIconKey(key string) (config.Item, bool) {
	cfg, _, err := s.store.Get()
	if err != nil {
		return config.Item{}, false
	}
	for _, sec := range cfg.Sections {
		for _, item := range sec.Items {
			if icons.ItemKey(item) == key {
				return item, true
			}
		}
	}
	return config.Item{}, false
}

func (s *Server) icon(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "__fallback" || !validKey(key) {
		writeFallback(w)
		return
	}
	data, ct, ok := s.icons.Read(key)
	if !ok {
		if item, found := s.itemByIconKey(key); found {
			s.icons.ForgetFail(key)
			if err := s.icons.Resolve(item); err == nil {
				data, ct, ok = s.icons.Read(key)
			}
		}
	}
	if !ok {
		writeFallback(w)
		return
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(data)
}

func validKey(key string) bool {
	if key == "" || len(key) > 120 {
		return false
	}
	for _, r := range key {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			continue
		}
		return false
	}
	if strings.Contains(key, "..") {
		return false
	}
	return true
}

func writeFallback(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = w.Write(icons.FallbackSVG())
}
