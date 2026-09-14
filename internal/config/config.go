package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	ErrItemNotFound    = errors.New("item not found")
	ErrItemExists      = errors.New("item already exists")
	ErrSectionNotFound = errors.New("section not found")
	ErrSectionExists   = errors.New("section already exists")
	ErrInvalidLayout   = errors.New("invalid layout")
)

type LayoutItem struct {
	URL string
	Col int
	Row int
}

type LayoutSection struct {
	Name  string
	Items []LayoutItem
}

type Config struct {
	Title    string    `json:"title"`
	Sections []Section `json:"sections"`
}

type Section struct {
	Name  string `json:"name"`
	Items []Item `json:"items"`
}

type Item struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Icon string `json:"icon,omitempty"`
	Col  int    `json:"col,omitempty"`
	Row  int    `json:"row,omitempty"`
}

type Store struct {
	path string

	mu    sync.RWMutex
	cfg   *Config
	mtime time.Time
}

func LoadFile(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if cfg.Title == "" {
		cfg.Title = "Lantern"
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	seen := map[string]struct{}{}
	for si, sec := range c.Sections {
		if sec.Name == "" {
			return fmt.Errorf("sections[%d]: name is required", si)
		}
		key := strings.ToLower(sec.Name)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("sections[%d]: duplicate name %q", si, sec.Name)
		}
		seen[key] = struct{}{}
		cells := map[string]struct{}{}
		for ii, item := range sec.Items {
			if item.Name == "" {
				return fmt.Errorf("sections[%d].items[%d]: name is required", si, ii)
			}
			if item.URL == "" {
				return fmt.Errorf("sections[%d].items[%d]: url is required", si, ii)
			}
			u, err := url.Parse(item.URL)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
				return fmt.Errorf("sections[%d].items[%d]: url must be http(s) with a host", si, ii)
			}
			if item.Icon != "" {
				if err := validateIcon(item.Icon); err != nil {
					return fmt.Errorf("sections[%d].items[%d]: %w", si, ii, err)
				}
			}
			if (item.Col > 0) != (item.Row > 0) {
				return fmt.Errorf("sections[%d].items[%d]: col and row must both be set", si, ii)
			}
			if item.Col < 0 || item.Row < 0 {
				return fmt.Errorf("sections[%d].items[%d]: col and row must be positive", si, ii)
			}
			if ItemHasPosition(item) {
				cell := fmt.Sprintf("%d:%d", item.Col, item.Row)
				if _, dup := cells[cell]; dup {
					return fmt.Errorf("sections[%d].items[%d]: duplicate grid cell (%d, %d)", si, ii, item.Col, item.Row)
				}
				cells[cell] = struct{}{}
			}
		}
	}
	return nil
}

func validateIcon(icon string) error {
	u, err := url.Parse(icon)
	if err != nil {
		return fmt.Errorf("invalid icon")
	}
	if u.Scheme == "" {
		return nil
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("icon URL must be http(s)")
	}
	if u.Host == "" {
		return fmt.Errorf("icon URL must include a host")
	}
	return nil
}

func NewStore(path string) (*Store, error) {
	cfg, err := LoadFile(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return &Store{path: path, cfg: cfg, mtime: info.ModTime()}, nil
}

// Get returns the current config, reloading from disk when mtime changes.
func (s *Store) Get() (cfg *Config, reloaded bool, err error) {
	info, err := os.Stat(s.path)
	if err != nil {
		s.mu.RLock()
		defer s.mu.RUnlock()
		if s.cfg != nil {
			return cloneConfig(s.cfg), false, nil
		}
		return nil, false, err
	}

	s.mu.RLock()
	same := info.ModTime().Equal(s.mtime)
	s.mu.RUnlock()
	if same {
		s.mu.RLock()
		defer s.mu.RUnlock()
		return cloneConfig(s.cfg), false, nil
	}

	cfg, err = LoadFile(s.path)
	if err != nil {
		s.mu.RLock()
		defer s.mu.RUnlock()
		if s.cfg != nil {
			return cloneConfig(s.cfg), false, nil
		}
		return nil, false, err
	}

	s.mu.Lock()
	s.cfg = cfg
	s.mtime = info.ModTime()
	s.mu.Unlock()
	return cloneConfig(cfg), true, nil
}

func (s *Store) loadForWrite() (*Config, error) {
	cfg, err := LoadFile(s.path)
	if err != nil {
		if s.cfg == nil {
			return nil, err
		}
		return cloneConfig(s.cfg), nil
	}
	return cfg, nil
}

func (s *Store) UpdateItem(sectionName, originalURL string, next Item) error {
	if s == nil {
		return fmt.Errorf("store unavailable")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadForWrite()
	if err != nil {
		return err
	}

	found := false
	for i, sec := range cfg.Sections {
		if sec.Name != sectionName {
			continue
		}
		for j, item := range sec.Items {
			if item.URL != originalURL {
				continue
			}
			if itemURLTaken(cfg, next.URL, sectionName, originalURL) {
				return ErrItemExists
			}
			next.Col = item.Col
			next.Row = item.Row
			cfg.Sections[i].Items[j] = next
			found = true
			break
		}
	}
	if !found {
		return ErrItemNotFound
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	return s.writeLocked(cfg)
}

func (s *Store) AddItem(sectionName string, next Item) error {
	if s == nil {
		return fmt.Errorf("store unavailable")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadForWrite()
	if err != nil {
		return err
	}
	if itemURLTaken(cfg, next.URL, "", "") {
		return ErrItemExists
	}
	found := false
	for i, sec := range cfg.Sections {
		if sec.Name != sectionName {
			continue
		}
		if !ItemHasPosition(next) {
			next.Col, next.Row = PlaceItemAtEnd(sec.Items, DefaultGridCols)
		}
		cfg.Sections[i].Items = append(cfg.Sections[i].Items, next)
		found = true
		break
	}
	if !found {
		return ErrSectionNotFound
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	return s.writeLocked(cfg)
}

func (s *Store) DeleteItem(sectionName, itemURL string) error {
	if s == nil {
		return fmt.Errorf("store unavailable")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadForWrite()
	if err != nil {
		return err
	}
	found := false
	for i, sec := range cfg.Sections {
		if sec.Name != sectionName {
			continue
		}
		next := make([]Item, 0, len(sec.Items))
		for _, item := range sec.Items {
			if item.URL == itemURL {
				found = true
				continue
			}
			next = append(next, item)
		}
		cfg.Sections[i].Items = next
		break
	}
	if !found {
		return ErrItemNotFound
	}
	return s.writeLocked(cfg)
}

func itemURLTaken(cfg *Config, wantURL, skipSection, skipOriginalURL string) bool {
	if wantURL == "" {
		return false
	}
	for _, sec := range cfg.Sections {
		for _, item := range sec.Items {
			if skipOriginalURL != "" && sec.Name == skipSection && item.URL == skipOriginalURL {
				continue
			}
			if item.URL == wantURL {
				return true
			}
		}
	}
	return false
}

func (s *Store) AddSection(name string) error {
	if s == nil {
		return fmt.Errorf("store unavailable")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("section name is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadForWrite()
	if err != nil {
		return err
	}
	want := strings.ToLower(name)
	for _, sec := range cfg.Sections {
		if strings.ToLower(sec.Name) == want {
			return ErrSectionExists
		}
	}
	cfg.Sections = append(cfg.Sections, Section{Name: name, Items: []Item{}})
	if err := cfg.Validate(); err != nil {
		return err
	}
	return s.writeLocked(cfg)
}

func (s *Store) SetLayout(layout []LayoutSection) error {
	if s == nil {
		return fmt.Errorf("store unavailable")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadForWrite()
	if err != nil {
		return err
	}

	itemsByURL := make(map[string]Item, 16)
	sectionsByName := make(map[string]Section, len(cfg.Sections))
	for _, sec := range cfg.Sections {
		sectionsByName[sec.Name] = sec
		for _, item := range sec.Items {
			itemsByURL[item.URL] = item
		}
	}

	seenSections := make(map[string]struct{}, len(cfg.Sections))
	usedURLs := make(map[string]struct{}, len(itemsByURL))
	next := make([]Section, 0, len(layout))
	for _, entry := range layout {
		name := strings.TrimSpace(entry.Name)
		if name == "" {
			return ErrInvalidLayout
		}
		if _, dup := seenSections[name]; dup {
			return ErrInvalidLayout
		}
		if _, ok := sectionsByName[name]; !ok {
			return ErrSectionNotFound
		}
		seenSections[name] = struct{}{}

		sec := Section{Name: name, Items: make([]Item, 0, len(entry.Items))}
		for _, layoutItem := range entry.Items {
			url := strings.TrimSpace(layoutItem.URL)
			if url == "" {
				return ErrInvalidLayout
			}
			item, ok := itemsByURL[url]
			if !ok {
				return ErrItemNotFound
			}
			if _, dup := usedURLs[url]; dup {
				return ErrInvalidLayout
			}
			usedURLs[url] = struct{}{}
			if layoutItem.Col > 0 && layoutItem.Row > 0 {
				item.Col = layoutItem.Col
				item.Row = layoutItem.Row
			} else {
				item.Col = 0
				item.Row = 0
			}
			sec.Items = append(sec.Items, item)
		}
		next = append(next, sec)
	}
	if len(seenSections) != len(sectionsByName) || len(usedURLs) != len(itemsByURL) {
		return ErrInvalidLayout
	}

	cfg.Sections = next
	if err := cfg.Validate(); err != nil {
		return err
	}
	return s.writeLocked(cfg)
}

func (s *Store) DeleteSection(name string) error {
	if s == nil {
		return fmt.Errorf("store unavailable")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("section name is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadForWrite()
	if err != nil {
		return err
	}
	next := make([]Section, 0, len(cfg.Sections))
	found := false
	for _, sec := range cfg.Sections {
		if sec.Name == name {
			found = true
			continue
		}
		next = append(next, sec)
	}
	if !found {
		return ErrSectionNotFound
	}
	cfg.Sections = next
	if err := cfg.Validate(); err != nil {
		return err
	}
	return s.writeLocked(cfg)
}

func (s *Store) writeLocked(cfg *Config) error {
	raw, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	if err := os.WriteFile(s.path, raw, 0o644); err != nil {
		return err
	}
	info, err := os.Stat(s.path)
	if err != nil {
		s.cfg = cfg
		return nil
	}
	s.cfg = cfg
	s.mtime = info.ModTime()
	return nil
}

func cloneConfig(src *Config) *Config {
	if src == nil {
		return nil
	}
	out := *src
	out.Sections = make([]Section, len(src.Sections))
	for i, sec := range src.Sections {
		out.Sections[i] = sec
		out.Sections[i].Items = append([]Item(nil), sec.Items...)
	}
	return &out
}
