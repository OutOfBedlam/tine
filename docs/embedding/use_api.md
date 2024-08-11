# Embedding TINE using API

**Create a pipeline**

```go
pipeline, err := engine.New(engine.WithName("my_pipeline"))
```

**Set inputs of the pipeline**

```go
// Add inlet for cpu usage
conf := engine.NewConfig().Set("percpu", false).Set("interval", 3 * time.Second)
pipeline.AddInlet("cpu", psutil.CpuInlet(pipeline.Context().WithConfig(conf)))
```

**Set outputs of the pipeline**

```go
// Add outlet printing to stdout '-'
conf = engine.NewConfig().Set("path", "-").Set("decimal", 2)
pipeline.AddOutlet("file", file.FileOutlet(pipeline.Context().WithConfig(conf)))
```

**Start the pipeline**

```go
pipeline.Start()
```
