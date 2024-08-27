package fakeit_test

import (
	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
	_ "github.com/OutOfBedlam/tine/plugins/fakeit"
)

func ExampleFakeItInlet() {
	// This example demonstrates how to use the fakeit inlet.
	dsl := `
	[[inlets.fakeit]]
		seed = 1
		count = 2
		interval = "1s"
		fields = [
			"name", "email", "phone", "city", "state", "zip", "country", 
			"latitude", "longitude", "int", "uint", "float", "unknown",
		]				
	[[outlets.file]]
		path = "-"
		format = "csv"
	`
	// Create a new engine.
	pipeline, err := engine.New(engine.WithConfig(dsl))
	if err != nil {
		panic(err)
	}
	// Run the engine.
	if err := pipeline.Run(); err != nil {
		panic(err)
	}
	// Output:
	// Sonny Stiedemann,codydonnelly@leannon.biz,7598907999,Tucson,Michigan,73793,United Arab Emirates,-34.489741,-1.754169,7824835832334018431,7714629679641582245,0.3503502922037831,unknown
	// Alfreda Doyle,luissmitham@wilderman.name,3591546690,Washington,Arizona,25709,Guinea,24.197279,49.312672,377978375824546181,14498895899673930012,0.24787804721849216,unknown
}
