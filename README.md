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

See the full code from the directory [./example/embed](./example/embed/main.go).

```go
// Create engine
eng, err := engine.New(engine.WithName("example"))
if err != nil {
    panic(err)
}

interval := 3 * time.Second
defaults := engine.NewConfig().Set("interval", interval)

// Add inlet for cpu usage
eng.AddInlet(
    psutil.CpuInlet(engine.NewConfig().Set("percpu", "false")), defaults)

// Add outlet printing to stdout '-'
eng.AddOutlet(
    file.FileOutlet(engine.NewConfig().Set("path", "-")), defaults)

// Add your custom input function.
custom := func() ([]engine.Record, error) {
    result := []engine.Record{
        engine.NewRecord(
            engine.NewStringField(engine.NAME_FIELD, "random"),
            engine.NewFloatField(engine.VALUE_FIELD, rand.Float64()*100),
        ),
    }
    return result, nil
}
eng.AddInlet(
    engine.InletWithFunc(custom, interval), defaults)

// Start the engine
go eng.Start()
```

Run this program, it shows the output like ...

```
cpu.percent,1720227446,7.1
random,1720227446,80.56
cpu.percent,1720227449,8.4
random,1720227449,83.61
cpu.percent,1720227452,8.9
random,1720227452,13.21
```

## Plugin system

### Format plugins

|  Name       |    Description |
|:------------|:---------------|
| `csv`       | [csv](./plugin/csv) encoder |
| `json`      | [json](./plugin/json) encoder |

### Compressor plugins

|  Name                           |    Description |
|:--------------------------------|:---------------|
| `gzip`, `zlib`, `lzw`, `flate`  | [compress](./plugin/compress) |
| `snappy`                        | [snappy](./plugin/snappy) encoder |


### Inlet plugins

> cpu,load,mem,disk,diskio,net,netstat,sensors,host,http,nats
  

### Outlet Plugins

> file,http,mqtt