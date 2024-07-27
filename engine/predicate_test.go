package engine_test

import (
	"fmt"

	"github.com/OutOfBedlam/tine/engine"
)

func ExampleAND() {
	// AND predicate
	// name == "John" AND age > 18
	p := engine.PredicateAND(
		engine.F{ColName: "name", Comparator: engine.EQ, Comparando: "John"},
		engine.F{ColName: "age", Comparator: engine.GT, Comparando: 18},
	)
	// Record
	records := []engine.Record{
		engine.NewRecord(
			engine.NewField("age", int64(19)),
			engine.NewField("name", "John"),
		),
		engine.NewRecord(
			engine.NewField("age", int64(17)),
			engine.NewField("name", "John"),
		),
	}
	// Apply AND predicate
	for _, r := range records {
		fmt.Println(p.Apply(r))
	}
	// Output:
	// true
	// false
}

func ExampleOR() {
	// OR predicate
	// name == "John" OR age < 18
	p := engine.PredicateOR(
		engine.F{ColName: "name", Comparator: engine.EQ, Comparando: "John"},
		engine.F{ColName: "age", Comparator: engine.LT, Comparando: 18},
	)
	// Record
	records := []engine.Record{
		engine.NewRecord(
			engine.NewField("age", int64(19)),
			engine.NewField("name", "John"),
		),
		engine.NewRecord(
			engine.NewField("age", int64(20)),
			engine.NewField("name", "Jane"),
		),
	}
	// Apply OR predicate
	for _, r := range records {
		fmt.Println(p.Apply(r))
	}
	// Output:
	// true
	// false
}

func ExampleNEQ() {
	// NEQ predicate
	// name != "John"
	p := engine.F{ColName: "name", Comparator: engine.NEQ, Comparando: "Jane"}
	// Record
	record := engine.NewRecord(engine.NewField("name", "John"))
	// Apply EQ predicate
	fmt.Println(p.Apply(record))
	// Output:
	// true
}

func ExampleIN() {
	// IN predicate
	// name IN ["John", "Jane", "Scott"]
	p := engine.F{ColName: "name", Comparator: engine.IN, Comparando: []string{"John", "Jane", "Scott"}}
	// Record
	record := []engine.Record{
		engine.NewRecord(engine.NewField("name", "John")),
		engine.NewRecord(engine.NewField("name", "Tiger")),
	}
	// Apply IN predicate
	for _, r := range record {
		fmt.Println(p.Apply(r))
	}
	// Output:
	// true
	// false
}

func ExampleIN_not() {
	// IN predicate
	// name IN ["John", "Jane", "Scott"]
	p := engine.F{ColName: "name", Comparator: engine.NOT_IN, Comparando: []string{"John", "Jane", "Scott"}}
	// Record
	record := []engine.Record{
		engine.NewRecord(engine.NewField("name", "John")),
		engine.NewRecord(engine.NewField("name", "Tiger")),
	}
	// Apply IN predicate
	for _, r := range record {
		fmt.Println(p.Apply(r))
	}
	// Output:
	// false
	// true
}

func ExampleCompFunc() {
	// EQ predicate
	// name == "John"
	p := engine.F{
		ColName:    "name",
		Comparator: engine.CompFunc,
		Comparando: func(v interface{}) bool {
			return v == "John"
		}}
	// Record
	record := engine.NewRecord(engine.NewField("name", "John"))
	// Apply EQ predicate
	fmt.Println(p.Apply(record))
	// Output:
	// true
}
