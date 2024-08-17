# Embedding TINE using API

**Create a pipeline**

```go
// import github.com/OutOfBedlam/tine/engine
//
// Create a pipeline
pipeline, err := engine.New(engine.WithName("my_pipeline"))
```

**Set inputs of the pipeline**

```go
// import github.com/OutOfBedlam/tine/plugins/psutil
//
// Add inlet for cpu usage
conf := engine.NewConfig().Set("percpu", false).Set("interval", 3 * time.Second)
pipeline.AddInlet("cpu", psutil.CpuInlet(pipeline.Context().WithConfig(conf)))
```

**Set outputs of the pipeline**

```go
// import github.com/OutOfBedlam/tine/plugins/base
//
// Add outlet printing to stdout '-'
conf = engine.NewConfig().Set("path", "-").Set("decimal", 2)
pipeline.AddOutlet("file", base.FileOutlet(pipeline.Context().WithConfig(conf)))
```

**Run the pipeline**

```go
pipeline.Run()
```
