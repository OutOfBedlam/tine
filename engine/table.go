package engine

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"sync"
)

type Key interface {
	cmp.Ordered
}

type Table[T Key] struct {
	mutex     sync.RWMutex
	columns   []string
	types     []Type
	rows      map[T]*Row[T]
	predicate Predicate
}

type Row[T Key] struct {
	Key    T
	Fields []*Field
}

func NewRow[T Key](key T, cap int) *Row[T] {
	return &Row[T]{
		Key:    key,
		Fields: make([]*Field, cap),
	}
}

func NewTable[T Key]() *Table[T] {
	return &Table[T]{
		rows: make(map[T]*Row[T]),
	}
}

func (tb *Table[T]) Keys() []T {
	keys := make([]T, 0, len(tb.rows))
	for k := range tb.rows {
		keys = append(keys, k)
	}
	if len(keys) == 0 {
		return nil
	}
	slices.Sort(keys)
	return keys
}

func (tb *Table[T]) Set(k T, fields ...*Field) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	var row *Row[T]
	if r, ok := tb.rows[k]; ok {
		row = r
	} else {
		row = NewRow(k, len(tb.columns))
		tb.rows[k] = row
	}

	var get_field = func(col string) *Field {
		for _, f := range fields {
			if f != nil && strings.EqualFold(f.Name, col) {
				return f
			}
		}
		return nil
	}

	for colIdx, col := range tb.columns {
		inField := get_field(col)
		if inField != nil {
			row.Fields[colIdx] = inField.Convert(tb.types[colIdx])
		}
	}

	for _, f := range fields {
		if f == nil {
			continue
		}
		colIdx := tb.columnIdx(f.Name)
		if colIdx >= 0 {
			continue
		}
		tb.AddColumn(f.Name, f.Type)
		row.Fields = append(row.Fields, f)
	}
	tb.rows[row.Key] = row
}

func (tb *Table[T]) Get(k T) *Row[T] {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return tb.rows[k]
}

func (tb *Table[T]) Rows() [][]*Field {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()

	keys := tb.Keys()
	ret := [][]*Field{}
	for _, k := range keys {
		row := tb.rows[k]
		ret = append(ret, row.Fields)
	}
	return ret
}

// AddColumns adds columns to the table
//
// names and types should be same size, otherwise panic
func (tb *Table[T]) AddColumns(names []string, types []Type) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	if len(names) != len(types) {
		panic("names and types should be same size")
	}
	for i := range names {
		tb.AddColumn(names[i], types[i])
	}
}

// AddColumn adds a column to the table
func (tb *Table[T]) AddColumn(name string, t Type) {
	tb.columns = append(tb.columns, strings.ToUpper(name))
	tb.types = append(tb.types, t)
}

// Columns returns the column names
func (tb *Table[T]) Columns() []string {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return tb.columns
}

// Types returns the column types
func (tb *Table[T]) Types() []Type {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return tb.types
}

// Clear removes all records from the table
func (tb *Table[T]) Clear() {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	clear(tb.rows)
}

// Len returns the number of records in the table
func (tb *Table[T]) Len() int {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return len(tb.rows)
}

// columnIdx returns the index of a column by name
func (tb *Table[T]) columnIdx(colName string) int {
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
func (tb *Table[T]) Series(colName string) []*Field {
	colIdx := tb.columnIdx(colName)
	if colIdx == -1 {
		return nil
	}
	return tb.SeriesByIdx(colIdx)
}

// SeriesByIdx returns a series of a column by index
func (tb *Table[T]) SeriesByIdx(colIdx int) []*Field {
	if colIdx < 0 || colIdx >= len(tb.columns) {
		return nil
	}
	keys := tb.Keys()
	ret := make([]*Field, len(keys))
	for i, k := range keys {
		ret[i] = tb.rows[k].Fields[colIdx]
	}
	return ret
}

// SeriesFields returns a series of a column by name
func (tb *Table[T]) SeriesFields(colName string) []*Field {
	colIdx := tb.columnIdx(colName)
	if colIdx == -1 {
		return nil
	}
	return tb.SeriesFieldsByIdx(colIdx)
}

// SeriesFieldsByIdx returns a series of a column by index
func (tb *Table[T]) SeriesFieldsByIdx(colIdx int) []*Field {
	if colIdx < 0 || colIdx >= len(tb.columns) {
		return nil
	}
	keys := tb.Keys()
	ret := make([]*Field, len(keys))
	for i, k := range keys {
		ret[i] = tb.rows[k].Fields[colIdx]
	}
	return ret
}

// Select returns a new table with selected columns
func (tb *Table[T]) Select(fields []string) (*Table[T], error) {
	colIndexes := make([]int, 0, len(fields))
	unknownCols := []string{}
	for _, col := range fields {
		idx := tb.columnIdx(col)
		if idx == -1 {
			unknownCols = append(unknownCols, col)
			continue
		}
		colIndexes = append(colIndexes, idx)
	}
	if len(unknownCols) > 0 {
		return nil, fmt.Errorf("unknown columns: %v", unknownCols)
	}
	ret := NewTable[T]()
	for _, idx := range colIndexes {
		ret.AddColumn(tb.columns[idx], tb.types[idx])
	}
	for k, row := range tb.rows {
		rec := sliceRecord(row.Fields)
		if r := rec.FieldsAt(colIndexes...); r != nil {
			if tb.predicate != nil && !tb.predicate.Apply(rec) {
				continue
			}
			ret.Set(k, r...)
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
func (tb *Table[T]) Compact() *Table[T] {
	if tb.predicate == nil {
		return tb
	}
	ret := &Table[T]{
		columns: tb.columns,
		types:   tb.types,
		rows:    make(map[T]*Row[T]),
	}
	for k, row := range tb.rows {
		rec := sliceRecord(row.Fields)
		if tb.predicate.Apply(rec) {
			ret.Set(k, row.Fields...)
		}
	}
	return ret
}

// Filter returns a new table with filtered records
func (tb *Table[T]) Filter(filter Predicate) *Table[T] {
	ret := &Table[T]{
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

func (tb *Table[T]) Split(filter Predicate) (*Table[T], *Table[T]) {
	ret := &Table[T]{
		columns: tb.columns,
		types:   tb.types,
		rows:    make(map[T]*Row[T]),
	}
	other := &Table[T]{
		columns: tb.columns,
		types:   tb.types,
		rows:    make(map[T]*Row[T]),
	}
	for k, row := range tb.rows {
		rec := sliceRecord(row.Fields)
		if filter.Apply(rec) {
			ret.Set(k, row.Fields...)
		} else {
			other.Set(k, row.Fields...)
		}
	}
	return ret, other
}
