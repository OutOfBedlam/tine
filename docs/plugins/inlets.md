# Inlets

### ARGS

Parse the command line arguments that following the `--` double dash. It assumes all arguments after the `--` are formed as key=value paris. And it pass them to the next step as a record.

For example, if the command line arguments are `tine some.toml -- --key1 value1 --key2 value2` then the record passed to the next step will be `{key1="value1", key2="value2"}`.

If value has prefix `base64+` followed by `http://`, `https://`, or `file://`, then the value is base64 encoded string of the content that are fetched from the URL or file.

If value has prefix `binary+` followed by `http://`, `https://`, or `file://`, then a `BinaryField` will be added instead of `StringField` within content that are fetched from the URL or file.

*Source* [plugins/args](https://github.com/OutOfBedlam/tine/tree/main/plugins/args)

**Config**

```toml
[[inlets.args]]
```

**Example**

```toml
[[inlets.args]]
[[outlets.file]]
    format = "json"
```

*Run*

```sh
tine run example.toml -- hello=world test=values
```

*Output*

```json
{"hello":"world","test":"values"}
```

### CPU

*Source* [plugins/psutil](https://github.com/OutOfBedlam/tine/tree/main/plugins/psutil)

**Config**

```toml
[[inlets.cpu]]
    percpu = false
    totalcpu = true
```

**Example**

```toml
[[inlets.cpu]]
    percpu = true
    totalcpu = true
    interval = "3s"
[[flows.select]]
    includes = ["**"]
[[outlets.file]]
    format = "json"
    decimal = 3
```

*Run*

```sh
tine run example.toml
```

*Output*

```json
{
    "0_percent":9.2,
    "1_percent":10.3,
    "2_percent":15.0,
    "3_percent":13.7,
    "_in":"cpu",
    "_ts":1722987663,
    "total_percent":9.6
}
```

### DISK

*Source* [plugins/psutil](https://github.com/OutOfBedlam/tine/tree/main/plugins/psutil)

**Config**

```toml
[[inlets.disk]]
    # default is all mount points
    mount_points = ["/", "/mnt"]
    ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "aufs", "squashfs"]

```

**Example**

```toml
[[inlets.disk]]
    mount_points = ["/"]
    interval = "3s"
[[flows.select]]
    includes = ["**"]
[[outlets.file]]
    format = "json"
    decimal = 0
```

*Run*

```sh
tine run example.toml
```

*Output*

```json
{
    "_in":"disk",
    "_ts":1722755225,
    "device":"/dev/disk3s1s1",
    "free":334269898752,
    "fstype":"apfs",
    "inodes_free":3264354480,
    "inodes_total":3264758647,
    "inodes_used":404167,
    "inodes_used_percent":0,
    "mount_point":"/",
    "total":994662584320,
    "used":660392685568,
    "used_percent":66
}
```

### DISKIO

*Source* [plugins/psutil](https://github.com/OutOfBedlam/tine/tree/main/plugins/psutil)

**Config**

```toml
[[inlets.diskio]]
    # default is all devices
    devices = ["sda*", "sdb*"]
```

**Example**

```toml
[[inlets.diskio]]
    devices = ["disk0"]
    interval = "3s"
[[flows.select]]
    includes = ["**"]
[[outlets.file]]
    format = "json"
```

*Run*

```sh
tine run example.toml
```

*Output*

```json
{
    "_in":"diskio",
    "_ts":1722755372,
    "device":"disk0",
    "io_time":11680599,
    "iops_in_progress":0,
    "label":"",
    "merged_read_count":0,
    "merged_write_count":0,
    "read_bytes":752462053376,
    "read_count":70590345,
    "read_time":10718681,
    "serial_number":"",
    "weighted_io":0,
    "write_bytes":178813861888,
    "write_count":9640875,
    "write_time":961918
}
```

### EXEC

Execute external command and yields records for the output of stdout of the command.

*Source* [plugins/exec](https://github.com/OutOfBedlam/tine/tree/main/plugins/exec)

**Config**

```toml
[[inlets.exec]]
    ## Commands array
    commands = ["uname", "-m"]

    ## Environment variables
    ## Array of "key=value" pairs
    ## e.g. ["key1=value1", "key2=value2"]
    environments = ["FOO=BAR"]

    ## Timeout
    timeout = "3s"

    ## Ignore non-zero exit code
    ignore_error = false

    ## Interval
    interval = "10s"

    ## How many times to run the command, 0 for infinite
    count = 0

    ## Trim space of output
    trim_space = false

    ## Separator for splitting output
    separator = ""

    ## Field name for stdout
    stdout_field = "stdout"

    ## Field name for stderr
    stderr_field = "stderr"
```

**Example**

```toml
[[inlets.exec]]
    commands = ["date", "+%s"]
    interval = "3s"
    count = 3
    trim_space = true
[[flows.select]]
    includes = ["#_in", "#_ts", "*"]
[[outlets.file]]
    format = "json"
```

*Output*

```json
{"_in":"exec","_ts":1722516453,"stdout":"1722516453"}
{"_in":"exec","_ts":1722516456,"stdout":"1722516456"}
{"_in":"exec","_ts":1722516459,"stdout":"1722516459"}
```

### FILE

*Source* [plugins/base](https://github.com/OutOfBedlam/tine/tree/main/plugins/base)

**Config**

```toml
[[inlets.file]]
    ### input file path, "-" for stdin, if "data" is specified, this field is ignored
    path = "/data/input.csv"
    ### input data in form of a string, if specified, "path" is ignored
    data = [
        "some,data,can,be,here",
        "and,here",
    ]
    ### input format
    format = "csv"
    ### name of the fields in the input data
    fields = ["name", "time","value"]
    ### type of the fields in the input data, the number of fields and types should be equal.
    ### if fields and types are not specified, all fields are treated as strings.
    types  = ["string", "time", "int"]
    ### Is input data compressed
    compress = ""
    ### time format (default: s)
    ### s, ms, us, ns, Golang timeformat string")
    ### e.g. timeformat = "2006-01-02 15:04:05 07:00"
    timeformat = "s"
    ### timezone (default: Local)
    ### e.g. tz = "Local"
    ### e.g. tz = "UTC"
    ### e.g. tz = "America/New_York"
    tz = "Local"
```

**Example**

```toml
[[inlets.file]]
    data = [
        "1,key1,1722642405,1.234",
        "2,key2,1722642406,2.345",
    ]
    format = "csv"
    fields = ["line", "name", "time", "value"]
    types  = ["int", "string", "time", "float"]
[[outlets.file]]
    format = "json"
```

*Run*

```sh
tine run example.toml
```

*Output*

```json
{"line":1,"name":"key1","time":1722642405,"value":1.234}
{"line":2,"name":"key2","time":1722642406,"value":2.345}
```

### HOST

*Source* [plugins/psutil](https://github.com/OutOfBedlam/tine/tree/main/plugins/psutil)

**Config**

```toml
[[inlets.host]]
```

**Example**

```toml
[[inlets.host]]
    interval = "3s"
[[flows.select]]
    includes = ["**"]
[[outlets.file]]
    format = "json"
```

*Run*

```sh
tine run example.toml
```

*Output*

```json
{
    "_in":"host",
    "_ts":1722755502,
    "host_id":"b4995381-ab81-5073-852f-d28b04a3ca4f",
    "hostname":"localhost.local",
    "kernel_arch":"arm64",
    "kernel_version":"23.5.0",
    "os":"darwin",
    "platform":"darwin",
    "platform_family":"Standalone Workstation",
    "platform_version":"14.5",
    "procs":746,
    "uptime":756732,
    "virtualization_role":"",
    "virtualization_system":""
}
```

### HTTP

*Source* [plugins/http](https://github.com/OutOfBedlam/tine/tree/main/plugins/http)

**Config**

```toml
[[inlets.http]]
    ### address e.g. http://localhost:8080
    address = ""

    ### success code (default: 200)
    success = 200

    ### timeout (default: 3s)
    timeout = "3s"

    interval = "10s"
    ### run count limit
    count = 1
```

**Example**

```toml
[[inlets.http]]
    address = "http://127.0.0.1:5555"
    success = 200
    timeout = "3s"
    count = 1
[[flows.select]]
    includes = ["**"]
[[outlets.file]]
    format = "json"
```

*Run*

```sh
tine run example.toml
```

*Output*

If the http server responded in JSON `{"a":1, "b":{"c":true, "d":3.14}}`.

The pipeline result will be:

```json
{"_in":"http","_ts":1721954797,"a":1,"b.c":true,"b.d":3.14}
```

### LOAD

*Source* [plugins/psutil](https://github.com/OutOfBedlam/tine/tree/main/plugins/psutil)

**Config**

```toml
[[inlets.load]]
    loads = [1, 5, 15]
```

**Example**

```toml
[[inlets.load]]
    loads = [1, 5, 15]
    interval = "3s"
[[flows.select]]
    includes = ["**"]
[[outlets.file]]
    format = "json"
    decimal = 2
```

*Run*

```sh
tine run example.toml
```

*Output*

```json
{
    "_in":"load",
    "_ts":1722987721,
    "load1":0.22,
    "load15":0.30,
    "load5":0.28
}
```

### MEM

*Source* [plugins/psutil](https://github.com/OutOfBedlam/tine/tree/main/plugins/psutil)

**Config**

```toml
[[inlets.mem]]
```

**Example**

```toml
[[inlets.mem]]
    interval = "3s"
[[flows.select]]
    includes = ["**"]
[[outlets.file]]
    format = "json"
    decimal = 2
```

*Run*

```sh
tine run example.toml
```

*Output*

```json
{
    "_in":"mem",
    "_ts":1722987763,
    "free":4330717184,
    "total":8201994240,
    "used":2468110336,
    "used_percent":30.09
}
```

### NATS_VARZ

*Source* [plugins/nats](https://github.com/OutOfBedlam/tine/tree/main/plugins/nats)

**Config**

```toml
[[inlets.nats_varz]]
    ### nats server statz endpoint
    server = "http://localhost:8222/varz"

    ### timeout (default: 3s)
    timeout = "3s"

    interval = "5s"
```

**Example**

```toml
[[inlets.nats_varz]]
    server = "http://localhost:8222/varz"
    timeout = "3s"
    interval = "5s"
[[flows.select]]
    includes = ["**"]
[[outlets.file]]
    format = "json"
    decimal = 2
```

*Run*

```sh
tine run ./example.toml
```

*Output*

```json
{
  "_in": "nats_varz",
  "_ts": 1723854808,
  "connections": 0,
  "cores": 10,
  "cpu": 0.00,
  "gomaxprocs": 10,
  "host": "0.0.0.0",
  "in_bytes": 0,
  "in_msgs": 0,
  "mem": 19378176,
  "out_bytes": 0,
  "out_msgs": 0,
  "port": 4222,
  "server_id": "NB5CLZPP2KHAOV6",
  "server_name": "NB5CLZPP2KHAOV6",
  "slow_consumers": 0,
  "subscriptions": 57,
  "total_connections": 0,
  "uptime": 1812,
  "version": "2.10.18"
}
```

### NET

*Source* [plugins/psutil](https://github.com/OutOfBedlam/tine/tree/main/plugins/psutil)

**Config**

```toml
[[inlets.net]]
    devices = ["eth*"]
```

**Example**

```toml
[[inlets.net]]
    devices = ["en0"]
    interval = "3s"
[[flows.select]]
    includes = ["**"]
[[outlets.file]]
    format = "json"
```

*Run*

```sh
tine run example.toml
```

*Output*

```json
{
    "_in":"net",
    "_ts":1722755428,
    "bytes_recv":6124558288,
    "bytes_sent":947180141,
    "device":"en0",
    "drop_in":0,
    "drop_out":11,
    "err_in":0,
    "err_out":0,
    "fifos_in":0,
    "fifos_out":0,
    "packets_recv":7065267,
    "packets_sent":0
}
```

### NETSTAT

*Source* [plugins/psutil](https://github.com/OutOfBedlam/tine/tree/main/plugins/psutil)

**Config**

```toml
[[inlets.netstat]]
    # macOS does not support netstat
    # supported protocols: ip, icmp, icmpmsg, tcp, udp, udplite
    protocols = ["tcp", "udp"]
```

**Example**

```toml
[[inlets.netstat]]
    protocols = ["tcp"]
[[flows.select]]
    includes = ["**"]
[[outlets.file]]
    format = "json"
```

*Run*

```sh
tine run ./example.toml
```

*Output*

```json
{
    "_in":"netstat",
    "_ts":1723855057,
    "active_opens":1447,
    "attempt_fails":0,
    "curr_estab":6,
    "estab_resets":0,
    "in_csum_errors":0,
    "in_errs":0,
    "in_segs":467153,
    "max_conn":-1,
    "out_rsts":11,
    "out_segs":469206,
    "passive_opens":5,
    "protocol":"tcp",
    "retrans_segs":666,
    "rto_algorithm":1,
    "rto_max":120000,
    "rto_min":200
}
```

### SCREENSHOT

*Source* [plugins/screenshot](https://github.com/OutOfBedlam/tine/tree/main/plugins/screenshot)

**Config**

```toml
[[inlets.screenshot]]
    ## The interval to run
    interval = "3s"
    ## How many times run before stop (0 is infinite)
    count = 1
    ## The display number to capture
    ## if it is empty, it will capture all displays
    displays = [0, 1]
    ## capture image format
    ## rgba | png | jpeg | gif
    format = "png"
```

**Example**

```toml
```

*Run*

```sh
```

*Output*

```json
```

### SENSORS

*Source* [plugins/psutil](https://github.com/OutOfBedlam/tine/tree/main/plugins/psutil)

**Config**

```toml
[[inlets.sensors]]
    interval = "10s"
```

**Example**

```toml
[[inlets.sensors]]
    interval = "5s"
[[flows.select]]
    includes = ["**"]
[[outlets.file]]
    format = "json"
    decimal = 2
```

*Run*

```sh
tine run example.toml
```

*Output*

```json
{
    "_in":"sensors",
    "_ts":1723670753,
    "critical":0.00,
    "high":0.00,
    "sensor_key":"PMU",
    "temperature":35.98
}
```

### SQLITE

*Source* [plugins/sqlite](https://github.com/OutOfBedlam/tine/tree/main/plugins/sqlite)

**Config**

```toml
```

**Example**

```toml
```

*Run*

```sh
```

*Output*

```json
```

### TELEGRAM

*Source* [plugins/telegram](https://github.com/OutOfBedlam/tine/tree/main/plugins/telegram)

**Config**

```toml
[[inlets.telegram]]
    token = "<bot_token>"
    debug = false
    timeout = "3s"
```

**Example**

```toml
```

*Run*

```sh
```

*Output*

```json
```
