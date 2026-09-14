package config

import "testing"

func TestResolvePositionsExplicitAndAuto(t *testing.T) {
	items := []Item{
		{Name: "A", URL: "http://a.lan", Col: 1, Row: 1},
		{Name: "B", URL: "http://b.lan"},
		{Name: "C", URL: "http://c.lan", Col: 3, Row: 2},
		{Name: "D", URL: "http://d.lan"},
	}
	got := ResolvePositions(items, 4)
	cases := []struct {
		url     string
		wantCol int
		wantRow int
	}{
		{"http://a.lan", 1, 1},
		{"http://b.lan", 2, 1},
		{"http://c.lan", 3, 2},
		{"http://d.lan", 4, 2},
	}
	for _, c := range cases {
		var item Item
		for _, entry := range got {
			if entry.URL == c.url {
				item = entry
				break
			}
		}
		if item.Col != c.wantCol || item.Row != c.wantRow {
			t.Fatalf("%s: got (%d,%d) want (%d,%d)", c.url, item.Col, item.Row, c.wantCol, c.wantRow)
		}
	}
}

func TestResolvePositionsSequentialWhenNoExplicit(t *testing.T) {
	items := []Item{
		{Name: "A", URL: "http://a.lan"},
		{Name: "B", URL: "http://b.lan"},
		{Name: "C", URL: "http://c.lan"},
	}
	got := ResolvePositions(items, 4)
	want := [][]int{{1, 1}, {2, 1}, {3, 1}}
	for i, item := range got {
		if item.Col != want[i][0] || item.Row != want[i][1] {
			t.Fatalf("item %d: got (%d,%d) want (%d,%d)", i, item.Col, item.Row, want[i][0], want[i][1])
		}
	}
}

func TestPlaceItemAtEnd(t *testing.T) {
	items := []Item{
		{Name: "A", URL: "http://a.lan", Col: 1, Row: 1},
		{Name: "B", URL: "http://b.lan", Col: 3, Row: 2},
	}
	col, row := PlaceItemAtEnd(items, 4)
	if col != 4 || row != 2 {
		t.Fatalf("got (%d,%d) want (4,2)", col, row)
	}
}

func TestValidateDuplicateGridCell(t *testing.T) {
	path := writeTemp(t, `{
		"sections": [{
			"name": "A",
			"items": [
				{"name": "X", "url": "http://x.lan", "col": 1, "row": 1},
				{"name": "Y", "url": "http://y.lan", "col": 1, "row": 1}
			]
		}]
	}`)
	if _, err := LoadFile(path); err == nil {
		t.Fatal("expected duplicate cell error")
	}
}
