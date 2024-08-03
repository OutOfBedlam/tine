# Inlets

### ARGS

Parse the command line arguments that following the `--` double dash. It assumes all arguments after the `--` are formed as key=value paris. And it pass them to the next step as a record.

For example, if the command line arguments are `tine some.toml -- --key1 value1 --key2 value2` then the record passed to the next step will be `{key1="value1", key2="value2"}`.

If value has prefix `base64+` followed by `http://`, `https://`, or `file://`, then the value is base64 encoded string of the content that are fetched from the URL or file.

If value has prefix `binary+` followed by `http://`, `https://`, or `file://`, then a `BinaryField` will be added instead of `StringField` within content that are fetched from the URL or file.

*Source* [plugin/inlets/args](https://github.com/OutOfBedlam/tine/tree/main/plugin/inlets/args)

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
[{"hello":"world","test":"values"}]
```

### EXEC

Execute external command and yields records for the output of stdout of the command.

*Source* [plugin/inlets/exec](https://github.com/OutOfBedlam/tine/tree/main/plugin/inlets/exec)

**Config**

```toml
[[inlets.exec]]
    ## Commands array
    commands = ["uname", "-m"]

    ## Environment variables
    ## Array of "key=value" pairs
    ## e.g. ["key1=value1", "key2=value2"]
    environments = []

    ## Field name prefix
    prefix = ""

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
[{"_in":"exec","_ts":1722516453,"stdout":"1722516453"}]
[{"_in":"exec","_ts":1722516456,"stdout":"1722516456"}]
[{"_in":"exec","_ts":1722516459,"stdout":"1722516459"}]
```

### FILE

*Source* [plugin/inlets/file](https://github.com/OutOfBedlam/tine/tree/main/plugin/inlets/file)

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
    field_names = ["name", "time","value"]
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
    field_names = ["line", "name", "time", "value"]
[[outlets.file]]
    format = "json"
    indent = "  "
```

*Run*

```sh
tine run example.toml
```

*Output*

```json
[
  {
    "line": "1",
    "name": "key1",
    "time": "1722642405",
    "value": "1.234"
  },
  {
    "line": "2",
    "name": "key2",
    "time": "1722642406",
    "value": "2.345"
  }
]
```

### HTTP

*Source* [plugin/inlets/http](https://github.com/OutOfBedlam/tine/tree/main/plugin/inlets/http)

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

### NATS

*Source* [plugin/inlets/nats](https://github.com/OutOfBedlam/tine/tree/main/plugin/inlets/nats)

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

### CPU

*Source* [plugin/inlets/psutil](https://github.com/OutOfBedlam/tine/tree/main/plugin/inlets/psutil)

**Config**

```toml
[[inlets.cpu]]
    percpu = false
    totalcpu = true
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

### LOAD

*Source* [plugin/inlets/psutil](https://github.com/OutOfBedlam/tine/tree/main/plugin/inlets/psutil)

**Config**

```toml
[[inlets.load]]
    loads = [1, 5, 15]
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

### MEM

*Source* [plugin/inlets/psutil](https://github.com/OutOfBedlam/tine/tree/main/plugin/inlets/psutil)

**Config**

```toml
[[inlets.mem]]
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

### DISK

*Source* [plugin/inlets/psutil](https://github.com/OutOfBedlam/tine/tree/main/plugin/inlets/psutil)

**Config**

```toml
[[inlets.disk]]
    # default is all mount points
    mount_points = ["/", "/mnt"]
    ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "aufs", "squashfs"]

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

### DISKIO

*Source* [plugin/inlets/psutil](https://github.com/OutOfBedlam/tine/tree/main/plugin/inlets/psutil)

**Config**

```toml
[[inlets.diskio]]
    # default is all devices
    devices = ["sda*", "sdb*"]
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

### NET

*Source* [plugin/inlets/psutil](https://github.com/OutOfBedlam/tine/tree/main/plugin/inlets/psutil)

**Config**

```toml
[[inlets.net]]
    devices = ["eth*"]
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

### NETSTAT

*Source* [plugin/inlets/psutil](https://github.com/OutOfBedlam/tine/tree/main/plugin/inlets/psutil)

**Config**

```toml
[[inlets.netstat]]
    # macOS does not support netstat
    # supported protocols: ip, icmp, icmpmsg, tcp, udp, udplite
    protocols = ["tcp", "udp"]
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

### SENSOR

*Source* [plugin/inlets/psutil](https://github.com/OutOfBedlam/tine/tree/main/plugin/inlets/psutil)

**Config**

```toml
[[inlets.sensor]]
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

### HOST

*Source* [plugin/inlets/psutil](https://github.com/OutOfBedlam/tine/tree/main/plugin/inlets/psutil)

**Config**

```toml
[[inlets.host]]
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

### SCREENSHOT

*Source* [plugin/inlets/screenshot](https://github.com/OutOfBedlam/tine/tree/main/plugin/inlets/screenshot)

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

### SQLITE

{% hint style="danger" %}
WIP
{% endhint %}

*Source* [plugin/inlets/sqlite](https://github.com/OutOfBedlam/tine/tree/main/plugin/inlets/sqlite)

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

*Source* [plugin/inlets/telegram](https://github.com/OutOfBedlam/tine/tree/main/plugin/inlets/telegram)

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
