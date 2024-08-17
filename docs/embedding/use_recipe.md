# Embedding TINE using Recipe

**Imports**

```go
import (
    github.com/OutOfBedlam/tine/engine
    _ github.com/OutOfBedlam/tine/plugins/base
    _ github.com/OutOfBedlam/tine/plugins/psutil
)
```

Whenever add new type of inlets, flows, outlets and codec, it should be imported.
If this is too cumbersome, import all plugins at once.

```go
import (
    github.com/OutOfBedlam/tine/engine
    _ github.com/OutOfBedlam/tine/plugins/all
)
```

**Define pipeline**

```go
const pipelineRecipe = `
[[inlets.cpu]]
	interval = "3s"
[[flows.select]]
	includes = ["#_ts", "*"]
[[outlets.file]]
	format = "json"
	decimal = 2
`
```

**Create a pipeline**

```go
pipeline, err := engine.New(engine.WithConfig(pipelineRecipe))
```

**Start the pipeline**

```go
pipeline.Start()
```

The `Start()` function in the code snippet above initiates the pipeline execution but does not wait for it to complete. Instead, it spawns a goroutine by calling `go pipeline.Run()` and returns immediately. On the other hand, `pipeline.Run()` is a blocking function that waits until the pipeline finishes its execution before returning control.

