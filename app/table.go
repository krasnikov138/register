package app

type Table[T any] struct {
	Values  [][]T
	Columns []string
}

func NewEmptyTable[T any]() *Table[T] {
	return &Table[T]{
		Values:  make([][]T, 0),
		Columns: make([]string, 0),
	}
}

func NewTable[T any](rows int, columns int) *Table[T] {
	var table Table[T]

	table.Columns = make([]string, columns)
	table.Values = make([][]T, columns)

	for i := 0; i < columns; i += 1 {
		table.Values[i] = make([]T, rows)
	}

	return &table
}

func NewTableCols[T any](row int, columns []string) *Table[T] {
	table := NewTable[T](row, len(columns))
	copy(table.Columns, columns)
	return table
}

func (table *Table[T]) IsEmpty() bool {
	return len(table.Columns) == 0
}

func (table *Table[T]) NCols() int {
	return len(table.Columns)
}

func (table *Table[T]) NRows() int {
	if len(table.Columns) == 0 {
		return 0
	}
	return len(table.Values[0])
}

func (table *Table[T]) GetColumn(name string) []T {
	for i, val := range table.Columns {
		if val == name {
			return table.Values[i]
		}
	}
	return nil
}

func (table *Table[T]) GetColumnIdx(name string) int {
	for i, val := range table.Columns {
		if val == name {
			return i
		}
	}
	return -1
}

func (table *Table[T]) GetColumnIndexer() map[string]int {
	indexer := make(map[string]int)

	for i, col := range table.Columns {
		indexer[col] = i
	}
	return indexer
}

func (table *Table[T]) GetRow(idx int) []T {
	result := make([]T, table.NCols())

	for i := range result {
		result[i] = table.Values[i][idx]
	}

	return result
}
