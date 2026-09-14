package config

const DefaultGridCols = 4

func ItemHasPosition(item Item) bool {
	return item.Col > 0 && item.Row > 0
}

func cellIndex(col, row, cols int) int {
	if cols < 1 {
		cols = 1
	}
	return (row-1)*cols + (col - 1)
}

func indexToCell(index, cols int) (col, row int) {
	if cols < 1 {
		cols = 1
	}
	if index < 0 {
		index = 0
	}
	return index%cols + 1, index/cols + 1
}

// ResolvePositions assigns grid cells in item order.
// Explicit col/row are kept (col clamped to cols).
// Items without positions are placed after the current end of the occupied area.
func ResolvePositions(items []Item, cols int) []Item {
	if len(items) == 0 {
		return items
	}
	if cols < 1 {
		cols = DefaultGridCols
	}

	out := make([]Item, len(items))
	maxIndex := -1
	for i, item := range items {
		out[i] = item
		if ItemHasPosition(item) {
			col := item.Col
			if col > cols {
				col = cols
			}
			out[i].Col = col
			out[i].Row = item.Row
			idx := cellIndex(col, item.Row, cols)
			if idx > maxIndex {
				maxIndex = idx
			}
			continue
		}
		nextIndex := maxIndex + 1
		col, row := indexToCell(nextIndex, cols)
		out[i].Col = col
		out[i].Row = row
		maxIndex = nextIndex
	}
	return out
}

// PlaceItemAtEnd assigns col/row for a new item at the end of the section grid.
func PlaceItemAtEnd(items []Item, cols int) (col, row int) {
	resolved := ResolvePositions(items, cols)
	maxIndex := -1
	for _, item := range resolved {
		idx := cellIndex(item.Col, item.Row, cols)
		if idx > maxIndex {
			maxIndex = idx
		}
	}
	return indexToCell(maxIndex+1, cols)
}
