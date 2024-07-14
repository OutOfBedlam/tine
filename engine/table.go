package engine

import (
	"fmt"
	"strings"
)

type Table struct {
	columns   []string
	types     []Type
	rows      [][]*Field
	predicate Predicate
}

func NewTable() *Table {
	return &Table{}
}

// AddColumns adds columns to the table
//
// names and types should be same size, otherwise panic
func (tb *Table) AddColumns(names []string, types []Type) {
	if len(names) != len(types) {
		panic("names and types should be same size")
	}
	for i := range names {
		tb.AddColumn(names[i], types[i])
	}
}

// AddColumn adds a column to the table
func (tb *Table) AddColumn(name string, t Type) {
	tb.columns = append(tb.columns, strings.ToUpper(name))
	tb.types = append(tb.types, t)
}

// Columns returns the column names
func (tb *Table) Columns() []string {
	return tb.columns
}

// Types returns the column types
func (tb *Table) Types() []Type {
	return tb.types
}

// Add adds a record to the table
func (tb *Table) Add(fields ...*Field) {
	rec := sliceRecord(fields)

	row := make([]*Field, len(tb.columns))
	for colIdx, colName := range tb.columns {
		field := rec.Field(colName)
		row[colIdx] = field.Convert(tb.types[colIdx])
	}

	tb.rows = append(tb.rows, row)
}

// Clear removes all records from the table
func (tb *Table) Clear() {
	tb.rows = tb.rows[:0]
}

// Len returns the number of records in the table
func (tb *Table) Len() int {
	return len(tb.rows)
}

// Row returns a record by index
func (tb *Table) Row(rowIdx int) []any {
	if rowIdx < 0 || rowIdx >= len(tb.rows) {
		return nil
	}
	ret := make([]any, len(tb.columns))
	for i, field := range tb.rows[rowIdx] {
		ret[i] = field.Value
	}
	return ret
}

// RowsFields returns a record by index
func (tb *Table) RowsFields(rowIdx int) []*Field {
	if rowIdx < 0 || rowIdx >= len(tb.rows) {
		return nil
	}
	return tb.rows[rowIdx]
}

// ColumnIdx returns the index of a column by name
func (tb *Table) ColumnIdx(colName string) int {
	colIdx := -1
	for i, name := range tb.columns {
		if name == strings.ToUpper(colName) {
			colIdx = i
			break
		}
	}
	return colIdx
}

// Series returns a series of a column by name
func (tb *Table) Series(colName string) []any {
	colIdx := tb.ColumnIdx(colName)
	if colIdx == -1 {
		return nil
	}
	return tb.SeriesByIdx(colIdx)
}

// SeriesByIdx returns a series of a column by index
func (tb *Table) SeriesByIdx(colIdx int) []any {
	if colIdx < 0 || colIdx >= len(tb.columns) {
		return nil
	}
	ret := make([]any, len(tb.rows))
	for i, row := range tb.rows {
		ret[i] = row[colIdx].Value
	}
	return ret
}

// SeriesFields returns a series of a column by name
func (tb *Table) SeriesFields(colName string) []*Field {
	colIdx := tb.ColumnIdx(colName)
	if colIdx == -1 {
		return nil
	}
	return tb.SeriesFieldsByIdx(colIdx)
}

// SeriesFieldsByIdx returns a series of a column by index
func (tb *Table) SeriesFieldsByIdx(colIdx int) []*Field {
	if colIdx < 0 || colIdx >= len(tb.columns) {
		return nil
	}
	ret := make([]*Field, len(tb.rows))
	for i, row := range tb.rows {
		ret[i] = row[colIdx]
	}
	return ret
}

// Select returns a new table with selected columns
func (tb *Table) Select(fields []string) (*Table, error) {
	colIndexes := make([]int, 0, len(fields))
	unknownCols := []string{}
	for _, col := range fields {
		idx := tb.ColumnIdx(col)
		if idx == -1 {
			unknownCols = append(unknownCols, col)
			continue
		}
		colIndexes = append(colIndexes, idx)
	}
	if len(unknownCols) > 0 {
		return nil, fmt.Errorf("unknown columns: %v", unknownCols)
	}
	ret := NewTable()
	for _, idx := range colIndexes {
		ret.AddColumn(tb.columns[idx], tb.types[idx])
	}
	for _, row := range tb.rows {
		rec := sliceRecord(row)
		if r := rec.FieldsAt(colIndexes...); r != nil {
			if tb.predicate != nil && !tb.predicate.Apply(rec) {
				continue
			}
			ret.Add(r...)
		}
	}
	return ret, nil
}

// Compact compacts records by predicate
// it removes records that do not match the predicate
// if predicate is nil, it does nothing.
// and returns the table itself. so that chain calls are possible.
//
// tb := tb.Filter(...).Compact().Select(...)
func (tb *Table) Compact() *Table {
	if tb.predicate == nil {
		return tb
	}
	rows := make([][]*Field, 0, len(tb.rows))
	for _, row := range tb.rows {
		rec := sliceRecord(row)
		if tb.predicate.Apply(rec) {
			rows = append(rows, row)
		}
	}
	tb.rows = rows
	return tb
}

// Filter returns a new table with filtered records
func (tb *Table) Filter(filter Predicate) *Table {
	ret := &Table{
		columns: tb.columns,
		types:   tb.types,
		rows:    tb.rows,
	}
	if ret.predicate != nil {
		ret.predicate = OR{tb.predicate, filter}
	} else {
		ret.predicate = filter
	}
	return ret
}
