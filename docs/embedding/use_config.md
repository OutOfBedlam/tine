# Embedding TINE using Config

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
const pipelineConfig = `
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
pipeline, err := engine.New(engine.WithConfig(pipelineConfig))
```

**Start the pipeline**

```go
pipeline.Start()
```

The `Start()` function does not wait for the pipeline to complete. Instead, it simply calls `go pipeline.Run()` and returns immediately.

