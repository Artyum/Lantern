package icons

import (
	"encoding/json"
	"strings"
)

const (
	maxCatalog            = 2 << 20
	simpleIconsCatalogURL = "https://cdn.jsdelivr.net/npm/simple-icons/data/simple-icons.json"
)

// simpleSlug is Simple Icons' titleToSlug: lowercase, a few substitutions,
// then ASCII letters and digits only. That is the CDN filename convention.
func simpleSlug(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.NewReplacer(
		"+", "plus",
		".", "dot",
		"&", "and",
		"đ", "d",
		"ħ", "h",
		"ı", "i",
		"ĸ", "k",
		"ŀ", "l",
		"ł", "l",
		"ß", "ss",
		"ŧ", "t",
		"ø", "o",
	).Replace(s)
	var b strings.Builder
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

type simpleIcon struct {
	Title   string         `json:"title"`
	Slug    string         `json:"slug"`
	Aliases *simpleAliases `json:"aliases"`
}

type simpleAliases struct {
	Aka []string          `json:"aka"`
	Old []string          `json:"old"`
	Loc map[string]string `json:"loc"`
	Dup []simpleDup       `json:"dup"`
}

type simpleDup struct {
	Title string `json:"title"`
}

func indexSimpleIcons(icons []simpleIcon) map[string]string {
	idx := make(map[string]string, len(icons)*2)
	add := func(key, slug string, overwrite bool) {
		key = simpleSlug(key)
		if key == "" || slug == "" {
			return
		}
		if !overwrite {
			if _, ok := idx[key]; ok {
				return
			}
		}
		idx[key] = slug
	}
	for _, ic := range icons {
		slug := ic.Slug
		if slug == "" {
			slug = simpleSlug(ic.Title)
		}
		add(slug, slug, true)
		add(ic.Title, slug, true)
	}
	for _, ic := range icons {
		slug := ic.Slug
		if slug == "" {
			slug = simpleSlug(ic.Title)
		}
		if ic.Aliases == nil {
			continue
		}
		for _, a := range ic.Aliases.Aka {
			add(a, slug, false)
		}
		for _, a := range ic.Aliases.Old {
			add(a, slug, false)
		}
		for _, a := range ic.Aliases.Loc {
			add(a, slug, false)
		}
		for _, d := range ic.Aliases.Dup {
			add(d.Title, slug, false)
		}
	}
	return idx
}

func (r *Resolver) catalogSlug(name string) string {
	r.ensureCatalog()
	key := simpleSlug(name)
	if key == "" || r.catalog == nil {
		return ""
	}
	return r.catalog[key]
}

func (r *Resolver) ensureCatalog() {
	r.catalogOnce.Do(func() {
		idx, err := r.loadCatalog()
		if err != nil {
			r.catalog = map[string]string{}
			return
		}
		r.catalog = idx
	})
}

func (r *Resolver) loadCatalog() (map[string]string, error) {
	body, _, err := r.fetchBytesN(simpleIconsCatalogURL, maxCatalog)
	if err != nil {
		return nil, err
	}
	var icons []simpleIcon
	if err := json.Unmarshal(body, &icons); err != nil {
		return nil, err
	}
	return indexSimpleIcons(icons), nil
}

func (r *Resolver) fetchSimpleFor(name string) ([]byte, string, error) {
	slug := simpleSlug(name)
	body, ext, err := r.fetchSimple(slug)
	if err == nil {
		return body, ext, nil
	}
	if mapped := r.catalogSlug(name); mapped != "" && mapped != slug {
		return r.fetchSimple(mapped)
	}
	return nil, "", err
}
