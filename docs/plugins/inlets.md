# Inlets

### ARGS

Parse the command line arguments that following the `--` double dash. It assumes all arguments after the `--` are formed as key=value paris. And it pass them to the next step as a record.

For example, if the command line arguments are `tine some.toml -- --key1 value1 --key2 value2` then the record passed to the next step will be `{key1="value1", key2="value2"}`.

If value has prefix `base64+` followed by `http://`, `https://`, or `file://`, then the value is base64 encoded string of the content that are fetched from the URL or file.

If value has prefix `binary+` followed by `http://`, `https://`, or `file://`, then a `BinaryField` will be added instead of `StringField` within content that are fetched from the URL or file.

*Package* `github.com/OutOfBedlam/tine/plugin/inlets/args`

#### Config

*N/A*

#### Example

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
[{"_in":"args","_ts":1722515124,"hello":"world","test":"values"}]
```

### EXEC

Execute external command and yields records for the output of stdout of the command.&#x20;

*Package* `github.com/OutOfBedlam/tine/plugin/inlets/exec`

#### Config

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

#### Example

```toml
[[inlets.exec]]
    commands = ["date", "+%s"]
    interval = "3s"
    count = 3
    trim_space = true
[[outlets.file]]
    format = "json"
```

*Output*

```json
[{"_in":"exec","_ts":1722515161,"stdout":"1722515161"}]
[{"_in":"exec","_ts":1722515164,"stdout":"1722515164"}]
[{"_in":"exec","_ts":1722515167,"stdout":"1722515167"}]
```

### FILE

### HTTP

### NATS

### CPU

### LOAD

### MEM

### DISK

### DISKIO

### NET

### NETSTAT

### SENSOR

### HOST

### SCREENSHOT

### SQLITE

### TELEGRAM

