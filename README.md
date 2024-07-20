# TINE

A straightforward data stream processor.

## Install

```bash
go install github.com/OutOfBedlam/tine
```

## Build

```bash
go build
```

## Usage

### Run

```bash
tine  <config.toml>
```

It generates metrics in CSV format with the following structure: `name, time, value`.

```
cpu.percent,1720158783,3.2
cpu.percent,1720158785,6.3
cpu.percent,1720158787,10.4
cpu.percent,1720158789,8.2
```

### Shebang

1. Save this file as `load.toml`

```toml
#!/path/to/tine
[log]
    level = "WARN"
[defaults]
    interval = "3s"
[[inlets.load]]
    loads = [1, 5]
[[outlets.file]]
    path  = "-"
```

2. Chmod for executable.

```sh
chmod +x load.toml
```

3. Run

```sh
$ ./load.tml

1720689517,load1,0.49
1720689517,load5,0.47
1720689517,load1,0.45
1720689517,load5,0.46
^C
```

## Embed in your program

Create an engine and add inlets and outlets.

See the full code from the directory [./example/custom_in](./example/helloworld/custom_in.go).

```go
// Create engine
pipeline, err := engine.New(engine.WithName("custom_in"))
if err != nil {
    panic(err)
}

interval := 3 * time.Second

// Add inlet for cpu usage
conf := engine.NewConfig().Set("percpu", false).Set("interval", interval)
pipeline.AddInlet("cpu", psutil.CpuInlet(pipeline.Context().WithConfig(conf)))

// Add outlet printing to stdout '-'
conf = engine.NewConfig().Set("path", "-").Set("decimal", 2)
pipeline.AddOutlet("file", file.FileOutlet(pipeline.Context().WithConfig(conf)))

// Add your custom input function.
custom := func() ([]engine.Record, error) {
    result := []engine.Record{
        engine.NewRecord(
            engine.NewStringField("name", "random"),
            engine.NewFloatField("value", rand.Float64()*100),
        ),
    }
    return result, nil
}
pipeline.AddInlet("custom", engine.InletWithPullFunc(custom, engine.WithInterval(interval)))

// Start the engine
go pipeline.Start()
```

Run this program, it shows the output like ...

```
1721510209,custom,random,43.01
1721510209,cpu,7.93
1721510212,custom,random,25.14
1721510212,cpu,7.35
1721510215,custom,random,83.24
1721510215,cpu,7.14
```

## Pipelines as your HTTP Handler

Use pipelines as your HTTP Handler, See the full code [./example/httpsvr](./example/httpsvr/httpsvr.go).

```go
var helloWorldPipeline = `
[[inlets.cpu]]
    interval = "3s"
    count = 1
    totalcpu = true
    percpu = true
[[outlets.file]]
    format = "json"
`
var screenshotPipeline = `
[[inlets.screenshot]]
    count = 1
    displays = [0]
[[outlets.image]]
    path = "nonamed.png"
`

router := http.NewServeMux()
router.HandleFunc("GET /helloworld", engine.HttpHandleFunc(helloWorldPipeline))
router.HandleFunc("GET /screenshot", engine.HttpHandleFunc(screenshotPipeline))
http.ListenAndServe(":8080", router)
```

## Plugin system

### Format plugins

|  Name       |    Description |
|:------------|:---------------|
| `csv`       | [csv](./plugin/codec/csv) encoder |
| `json`      | [json](./plugin/codec/json) encoder |

### Compressor plugins

|  Name                           |    Description |
|:--------------------------------|:---------------|
| `gzip`, `zlib`, `lzw`, `flate`  | [compress](./plugin/codec/compress) |
| `snappy`                        | [snappy](./plugin/codec/snappy) encoder |


### List plugins

```sh
$ tine --list

Input: data -> [inlet] -> [decompress] -> [decoder] -> records
  Decoders    csv
  Decompress  flate,gzip,inflate,lzw,snappy,zlib
  Inlets      exec,file,http,cpu,load,mem,disk,diskio,net,sensors,host,screenshot,syslog,rrd_graph,nats

Ouput: records -> [encoder] -> [compress] -> [outlet] -> data
  Encoders    csv,json
  Compress    deflate,flate,gzip,lzw,snappy,zlib
  Outlets     excel,file,http,image,rrd,mqtt

Flows         set_field_name,fan-in,fan-out,merge,flatten,damper,dump,set_field
```
