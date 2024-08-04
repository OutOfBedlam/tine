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
[[in.http]]
    ### address e.g. http://localhost:8080
    address = ""

    ### success code (default: 200)
    success = 200

    ### timeout (default: 3s)
    timeout = "3s"

    interval = "10s"
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
[[in.nats]]
    ### nats server statz endpoint
    server = ""

    ### timeout (default: 3s)
    timeout = "3s"

    interval = "10s"
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
[[inlets.cpu]]
    percpu = true
    totalcpu = true
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
[{"0_percent":33.207320361243234,"1_percent":32.53797979613833,"2_percent":6.87165938794736,"3_percent":3.8726567063664827,"_in":"cpu","_ts":1722755030,"total_percent":8.2044917374524}]
[{"0_percent":17.7474402743563,"1_percent":16.949152543944706,"2_percent":8.474576271152998,"3_percent":3.6666666662010052,"_in":"cpu","_ts":1722755033,"total_percent":5.130784707591248}]
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
[[inlets.load]]
    loads = [1, 5, 15]
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
[{"_in":"load","_ts":1722755103,"load1":1.55224609375,"load15":2.19921875,"load5":2.1376953125}]
[{"_in":"load","_ts":1722755106,"load1":1.5078125,"load15":2.19189453125,"load5":2.11865234375}]
```

### MEM

*Source* [plugin/inlets/psutil](https://github.com/OutOfBedlam/tine/tree/main/plugin/inlets/psutil)

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
```

*Run*

```sh
tine run example.toml
```

*Output*

```json
[{"_in":"mem","_ts":1722755153,"free":336723968,"total":34359738368,"used":20005076992,"used_percent":58.22243690490723}]
[{"_in":"mem","_ts":1722755156,"free":441221120,"total":34359738368,"used":19877855232,"used_percent":57.8521728515625}]
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
[[inlets.disk]]
    mount_points = ["/"]
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
[{"_in":"disk","_ts":1722755225,"device":"/dev/disk3s1s1","free":334269898752,"fstype":"apfs","inodes_free":3264354480,"inodes_total":3264758647,"inodes_used":404167,"inodes_used_percent":0.012379690007755724,"mount_point":"/","total":994662584320,"used":660392685568,"used_percent":66.39363900668654}]
[{"_in":"disk","_ts":1722755228,"device":"/dev/disk3s1s1","free":334269898752,"fstype":"apfs","inodes_free":3264354480,"inodes_total":3264758647,"inodes_used":404167,"inodes_used_percent":0.012379690007755724,"mount_point":"/","total":994662584320,"used":660392685568,"used_percent":66.39363900668654}]
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
[{"_in":"diskio","_ts":1722755372,"device":"disk0","io_time":11680599,"iops_in_progress":0,"label":"","merged_read_count":0,"merged_write_count":0,"read_bytes":752462053376,"read_count":70590345,"read_time":10718681,"serial_number":"","weighted_io":0,"write_bytes":178813861888,"write_count":9640875,"write_time":961918}]
[{"_in":"diskio","_ts":1722755375,"device":"disk0","io_time":11680599,"iops_in_progress":0,"label":"","merged_read_count":0,"merged_write_count":0,"read_bytes":752462053376,"read_count":70590345,"read_time":10718681,"serial_number":"","weighted_io":0,"write_bytes":178813861888,"write_count":9640875,"write_time":961918}]
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
[{"_in":"net","_ts":1722755428,"bytes_recv":6124558288,"bytes_sent":947180141,"device":"en0","drop_in":0,"drop_out":11,"err_in":0,"err_out":0,"fifos_in":0,"fifos_out":0,"packets_recv":7065267,"packets_sent":0}]
[{"_in":"net","_ts":1722755431,"bytes_recv":6124558853,"bytes_sent":947180654,"device":"en0","drop_in":0,"drop_out":11,"err_in":0,"err_out":0,"fifos_in":0,"fifos_out":0,"packets_recv":7065273,"packets_sent":0}]
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
    # macOS does not support sensor
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
[{"_in":"host","_ts":1722755502,"host_id":"b493ab81-9581-5523-807f-d28b04a3ca4f","hostname":"localhost.local","kernel_arch":"arm64","kernel_version":"23.5.0","os":"darwin","platform":"darwin","platform_family":"Standalone Workstation","platform_version":"14.5","procs":746,"uptime":756732,"virtualization_role":"","virtualization_system":""}]
[{"_in":"host","_ts":1722755505,"host_id":"b493ab81-9581-5523-807f-d28b04a3ca4f","hostname":"localhost.local","kernel_arch":"arm64","kernel_version":"23.5.0","os":"darwin","platform":"darwin","platform_family":"Standalone Workstation","platform_version":"14.5","procs":746,"uptime":756735,"virtualization_role":"","virtualization_system":""}]
```

### SCREENSHOT

*Source* [plugin/inlets/screenshot](https://github.com/OutOfBedlam/tine/tree/main/plugin/inlets/screenshot)

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
