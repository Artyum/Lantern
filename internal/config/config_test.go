package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTemp(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadFileValid(t *testing.T) {
	path := writeTemp(t, `{
		"title": "My LAN",
		"sections": [{
			"name": "Media",
			"items": [{
				"name": "Jellyfin",
				"url": "http://jellyfin.lan",
				"icon": "jellyfin"
			}]
		}]
	}`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Title != "My LAN" {
		t.Fatalf("title: %q", cfg.Title)
	}
	if len(cfg.Sections) != 1 || cfg.Sections[0].Items[0].Name != "Jellyfin" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestLoadFileDefaultTitle(t *testing.T) {
	path := writeTemp(t, `{"sections":[{"name":"A","items":[{"name":"B","url":"https://b.lan"}]}]}`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Title != "Lantern" {
		t.Fatalf("title: %q", cfg.Title)
	}
}

func TestValidateMissingFields(t *testing.T) {
	cases := []string{
		`{"sections":[{"name":"","items":[]}]}`,
		`{"sections":[{"name":"A","items":[{"name":"","url":"http://x.lan"}]}]}`,
		`{"sections":[{"name":"A","items":[{"name":"X","url":""}]}]}`,
		`{"sections":[{"name":"A","items":[{"name":"X","url":"ftp://x.lan"}]}]}`,
		`{"sections":[{"name":"A","items":[{"name":"X","url":"not-a-url"}]}]}`,
		`{"sections":[{"name":"A","items":[{"name":"X","url":"http://ok.lan","icon":"file:///etc/passwd"}]}]}`,
	}
	for _, body := range cases {
		path := writeTemp(t, body)
		if _, err := LoadFile(path); err == nil {
			t.Fatalf("expected error for %s", body)
		}
	}
}

func TestStoreReloadOnMtime(t *testing.T) {
	path := writeTemp(t, `{"title":"v1","sections":[{"name":"A","items":[{"name":"X","url":"http://x.lan"}]}]}`)
	store, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	cfg, reloaded, err := store.Get()
	if err != nil || reloaded || cfg.Title != "v1" {
		t.Fatalf("first get: cfg=%v reloaded=%v err=%v", cfg, reloaded, err)
	}

	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(path, []byte(`{"title":"v2","sections":[{"name":"A","items":[{"name":"X","url":"http://x.lan"}]}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, reloaded, err = store.Get()
	if err != nil || !reloaded || cfg.Title != "v2" {
		t.Fatalf("reload: cfg=%v reloaded=%v err=%v", cfg, reloaded, err)
	}
}

func TestUpdateItemPersists(t *testing.T) {
	path := writeTemp(t, `{"title":"v1","sections":[{"name":"A","items":[{"name":"X","url":"http://x.lan"}]}]}`)
	store, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	err = store.UpdateItem("A", "http://x.lan", Item{
		Name: "Y",
		URL:  "https://y.lan",
		Icon: "grafana",
	})
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	item := cfg.Sections[0].Items[0]
	if item.Name != "Y" || item.URL != "https://y.lan" || item.Icon != "grafana" {
		t.Fatalf("%+v", item)
	}
	if err := store.UpdateItem("A", "http://missing.lan", Item{Name: "Z", URL: "http://z.lan"}); !errors.Is(err, ErrItemNotFound) {
		t.Fatalf("missing: %v", err)
	}
}

func TestAddAndDeleteItem(t *testing.T) {
	path := writeTemp(t, `{"title":"v1","sections":[{"name":"A","items":[{"name":"X","url":"http://x.lan"}]}]}`)
	store, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.AddItem("A", Item{Name: "Y", URL: "http://y.lan", Icon: "grafana"}); err != nil {
		t.Fatal(err)
	}
	if err := store.AddItem("A", Item{Name: "Dup", URL: "http://x.lan"}); !errors.Is(err, ErrItemExists) {
		t.Fatalf("dup: %v", err)
	}
	if err := store.AddItem("Brak", Item{Name: "Z", URL: "http://z.lan"}); !errors.Is(err, ErrSectionNotFound) {
		t.Fatalf("section: %v", err)
	}
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Sections[0].Items) != 2 || cfg.Sections[0].Items[1].Name != "Y" {
		t.Fatalf("%+v", cfg.Sections[0].Items)
	}
	if err := store.DeleteItem("A", "http://x.lan"); err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteItem("A", "http://x.lan"); !errors.Is(err, ErrItemNotFound) {
		t.Fatalf("missing: %v", err)
	}
	cfg, err = LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Sections[0].Items) != 1 || cfg.Sections[0].Items[0].URL != "http://y.lan" {
		t.Fatalf("%+v", cfg.Sections[0].Items)
	}
}

func TestSetLayoutReordersSectionsAndItems(t *testing.T) {
	path := writeTemp(t, `{
		"title": "v1",
		"sections": [
			{"name": "A", "items": [
				{"name": "X", "url": "http://x.lan"},
				{"name": "Y", "url": "http://y.lan"}
			]},
			{"name": "B", "items": [
				{"name": "Z", "url": "http://z.lan"}
			]}
		]
	}`)
	store, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SetLayout([]LayoutSection{
		{Name: "B", Items: []string{"http://z.lan"}},
		{Name: "A", Items: []string{"http://y.lan", "http://x.lan"}},
	}); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Sections) != 2 || cfg.Sections[0].Name != "B" || cfg.Sections[1].Name != "A" {
		t.Fatalf("sections: %+v", cfg.Sections)
	}
	if cfg.Sections[1].Items[0].URL != "http://y.lan" || cfg.Sections[1].Items[1].URL != "http://x.lan" {
		t.Fatalf("items: %+v", cfg.Sections[1].Items)
	}
	if err := store.SetLayout([]LayoutSection{
		{Name: "A", Items: []string{"http://z.lan"}},
		{Name: "B", Items: []string{"http://x.lan", "http://y.lan"}},
	}); err != nil {
		t.Fatal(err)
	}
	cfg, err = LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Sections[0].Items[0].URL != "http://z.lan" || cfg.Sections[0].Name != "A" {
		t.Fatalf("moved item: %+v", cfg.Sections[0])
	}
}

func TestSetLayoutValidation(t *testing.T) {
	path := writeTemp(t, `{"title":"v1","sections":[{"name":"A","items":[{"name":"X","url":"http://x.lan"}]}]}`)
	store, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SetLayout([]LayoutSection{{Name: "A", Items: []string{"http://missing.lan"}}}); !errors.Is(err, ErrItemNotFound) {
		t.Fatalf("missing item: %v", err)
	}
	if err := store.SetLayout([]LayoutSection{{Name: "A", Items: []string{"http://x.lan", "http://x.lan"}}}); !errors.Is(err, ErrInvalidLayout) {
		t.Fatalf("duplicate item: %v", err)
	}
	if err := store.SetLayout([]LayoutSection{{Name: "Brak", Items: []string{}}}); !errors.Is(err, ErrSectionNotFound) {
		t.Fatalf("missing section: %v", err)
	}
}

func TestAddAndDeleteSection(t *testing.T) {
	path := writeTemp(t, `{"title":"v1","sections":[{"name":"A","items":[{"name":"X","url":"http://x.lan"}]}]}`)
	store, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.AddSection("Nowa"); err != nil {
		t.Fatal(err)
	}
	if err := store.AddSection("nowa"); !errors.Is(err, ErrSectionExists) {
		t.Fatalf("dup: %v", err)
	}
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Sections) != 2 || cfg.Sections[1].Name != "Nowa" {
		t.Fatalf("%+v", cfg.Sections)
	}
	if err := store.DeleteSection("A"); err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteSection("A"); !errors.Is(err, ErrSectionNotFound) {
		t.Fatalf("missing: %v", err)
	}
	cfg, err = LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Sections) != 1 || cfg.Sections[0].Name != "Nowa" {
		t.Fatalf("%+v", cfg.Sections)
	}
}
