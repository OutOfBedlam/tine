package expr_test

import (
	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
	_ "github.com/OutOfBedlam/tine/plugins/expr"
)

func ExampleMapFlow() {
	recipe := `
	[[inlets.file]]
		data = [
			"a,100,1.234,1724136681,true",
			"b,200,2.345,1724136682,false",
			"c,300,3.456,1724136683,true",
		]
		format = "csv"
		fields = ["name", "rec.value", "fval", "tval", "flag"]
		types  = ["string", "int", "float", "time", "bool"]
	[[flows.map]]
		code = "rec.value = ${ rec.value } * 2"
	[[flows.map]]
		code = "fval = ${ fval } * 2"
	[[flows.map]]
		code = "flag = !${ flag }"
	[[flows.map]]
		predicate = "${ name } == 'a' || ${name} == 'c'"
		code = "name = 'hello ' + ${ name }"
	[[flows.map]]
		code = '''printf("hello world %v\n", ${rec.value})'''
	[[outlets.file]]
		path = "-"
		format = "json"
`
	pipe, err := engine.New(engine.WithConfig(recipe))
	if err != nil {
		panic(err)
	}
	err = pipe.Run()
	if err != nil {
		panic(err)
	}

	// Output:
	// hello world 200
	// hello world 400
	// hello world 600
	// {"flag":false,"fval":2.468,"name":"hello a","rec.value":200,"tval":1724136681}
	// {"flag":true,"fval":4.69,"name":"b","rec.value":400,"tval":1724136682}
	// {"flag":false,"fval":6.912,"name":"hello c","rec.value":600,"tval":1724136683}
}
