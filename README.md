# TINE

![CI](https://github.com/OutOfBedlam/tine/actions/workflows/ci.yml/badge.svg)
[![codecov](https://codecov.io/gh/OutOfBedlam/tine/graph/badge.svg?token=5XSG9M9P8E)](https://codecov.io/gh/OutOfBedlam/tine)


![TINE is not ETL](./docs/images/tine-drop-circlex256.png)

TINE a data pipeline runner.

## Install

```bash
go install github.com/OutOfBedlam/tine
```

## Build

```bash
go build
```

## Usage

### Define pipeline in TOML

Set the pipline's inputs and outputs.

```toml
[log]
    level = "WARN"
[defaults]
    interval = "3s"
[[inlets.cpu]]
[[outlets.file]]
    path  = "-"
```

### Run

```bash
tine run <config.toml>
```

It generates CPU usage in CSV format which is default format of 'outlets.file'.

```
1721635296,cpu,1.5774180156295268
1721635299,cpu,0.6677796326153481
1721635302,cpu,1.079734219344818
1721635305,cpu,2.084201750601381
```

And change output format to "json" from "csv".

```toml
...omit...
[[outlets.file]]
    path  = "-"
    format = "json"
```

```json
[{"_in":"cpu","_ts":1721780188,"total_percent":0.9166666666362681}]
[{"_in":"cpu","_ts":1721780191,"total_percent":1.0403662089355488}]
[{"_in":"cpu","_ts":1721780194,"total_percent":0.2507312996272184}]
[{"_in":"cpu","_ts":1721780197,"total_percent":1.2093411175800368}]
```

### Shebang

1. Save this file as `load.toml`

```toml
#!/path/to/tine run
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
$ ./load.toml

1721635438,load,0.03,0.08
1721635441,load,0.03,0.08
1721635444,load,0.03,0.08
^C
```

## Embed in your program

Create an pipeline and add inlets and outlets.

See the full code from the directory [./example/custom_in](./example/helloworld/custom_in.go).

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

**Run the pipeline**

```
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

## Pipelines as HTTP Handler

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

## Examples & recipes

- See more [examples](./example/).

## Plugin system

**Inbound**

`data -> [inlet] -> [decompress] -> [decoder] -> records`

**Outbound**

`records -> [encoder] -> [compress] -> [outlet] -> data`

### Codec

|  Name       |    Description                      |
|:------------|:------------------------------------|
| `csv`       | [csv](./plugin/codec/csv) codec     |
| `json`      | [json](./plugin/codec/json) encoder only |

### Compressor

|  Name                           |    Description |
|:--------------------------------|:---------------|
| `gzip`, `zlib`, `lzw`, `flate`  | [compress](./plugin/codec/compress) |
| `snappy`                        | [snappy](./plugin/codec/snappy) encoder |


### Inlets

|  Name        |    Description                             |
|:-------------|:-------------------------------------------|
| `args`       | Generates a record from os.Args            |
| `exec`       | Execute external commands and reads stdout |
| `file`       | Read a file                                |
| `http`       | Retrieve from a http end point             |
| `cpu`        | CPU usage percent                          |
| `load`       | System load                                |
| `mem`        | System memory usage percent                |
| `disk`       | Disk usage percent                         |
| `diskio`     | Disk IO stat                               |
| `net`        | Network traffic stat                       |
| `sensors`    | System sensors                             |
| `host`       | Host stat                                  |
| `screenshot` | Take screenshot of desktop                 |
| `sqlite`     | sqlite query                               |
| `syslog`     | Receive rsyslog messages via network       |
| `telegram`   | Receive messages via Telegram              |
| `nats`       | NATS server stat                           |
| `rrd-graph`  | rrd graph (required rebuild `go build -tags rrd`) |

### Outlets

|  Name        |    Description                             |
|:-------------|:-------------------------------------------|
| `excel`      | Write am excel file                        |
| `file`       | Write to a file                            |
| `http`       | Post data to http endpoint                 |
| `image`      | Save image files                           |
| `sqlite`     | sqlite                                     |
| `template`   | Apply template and write to a file         |
| `telegram`   | Send message via Telegram                  |
| `mqtt`       | Publish to MQTT broker                     |
| `rrd`        | rrd (required rebuild `go build -tags rrd` ) |

### Flows

|  Name          |    Description                             |
|:---------------|:-------------------------------------------|
| set_field_name | Manipulate field name of records           |
| set_field      | Forcely set a field                        |
| fan-in         | Aggregate messages from multiple sources   |
| fan-out        | Distribute messages to multiple sinks      |
| merge          | Merge multiple records into a wide record  |
| flatten        | Split a record into multiple records       |
| damper         | Combine multiple records                   |
| dump           | Log print records for debugging            |

### Show plugin list

```sh
$ tine list

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
