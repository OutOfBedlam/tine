# Flows

### DAMPER

*Source* [plugins/base](https://github.com/OutOfBedlam/tine/tree/main/plugins/base)

**Config**

```toml
[[flows.damper]]
    ## damper makes the stream of records to be delayed by the given duration.
    ## It collects records those _ts time is older thant now-"interval" time, 
    ## and send them to the next flow in a time.
    interval = "3s"
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

### DUMP

*Source* [plugins/base](https://github.com/OutOfBedlam/tine/tree/main/plugins/base)

**Config**

```toml
[[flows.dump]]
    ## dump writes the record to the log with the given log level.
    ## DEBUG | INFO | WARN | ERROR  (default: DEBUG)
	level = "DEBUG"
    ## The decimal format for float fields. (default: -1 which means no rounding)
    decimal = 2
    ## The time format for time fields. (default: "2006-01-02 15:04:05")
    timeformat = "2006-01-02 15:04:05"
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

### EXEC

*Source* [plugins/exec](https://github.com/OutOfBedlam/tine/tree/main/plugins/exec)

**Config**

```toml
[[flows.exec]]
    ## Commands array
    ## Access the field and tag values in the command by using 
    ## the uppercase field name prefixed with "$FIELD_{name}"
    ## and uppercase tag name prefixed with "$TAG_{name}"
    commands = ["echo", "$FOO", "$FIELD_SOME", "$TAG__IN"]

    ## Environment variables
    ## Array of "key=value" pairs
    ## e.g. ["key1=value1", "key2=value2"]
    environments = ["FOO=BAR"]

    ## Timeout
    timeout = "3s"

    ## Ignore non-zero exit code
    ignore_error = false

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
[[inlets.file]]
    data = [
        "a,1",
        "b,2",
    ]
    format = "csv"
[[flows.exec]]
    commands = ["sh", "-c", "echo hello $FOO $FIELD_0 $FIELD_1"]
    environments = ["FOO=BAR"]
    trim_space = true
    ignore_error = true
    stdout_field = "output
[[flows.select]]
    includes= ["#_ts", "*"]
[[outlets.file]]
    path = "-"
    format = "json"
```

*Run*

```sh
tine run example.toml
```

*Output*

```json
{"_ts":1721954798,"output":"hello BAR a 1"}
{"_ts":1721954799,"output":"hello BAR b 2"}
```

### MERGE

*Source* [plugins/base](https://github.com/OutOfBedlam/tine/tree/main/plugins/base)

**Config**

```toml
[[flows.merge]]
    wait_limit = "2s"
```

**Example**

```toml
[[inlets.cpu]]
    percpu = false
    interval = "1s"
    count = 3
[[inlets.load]]
    loads = [1, 5]
    interval = "1s"
    count = 2
[[flows.merge]]
    wait_limit = "1s"
[[outlets.file]]
    path = "-"
    format = "json"
    decimal = 2
```

*Run*

```sh
tine run example.toml
```

*Output*

```json
{"_ts":1723248243,"cpu.total_percent":8.16,"load.load1":1.90,"load.load5":1.94}
{"_ts":1723248244,"cpu.total_percent":11.67,"load.load1":1.90,"load.load5":1.94}
{"_ts":1723248245,"cpu.total_percent":15.56}
```

### UPDATE

*Source* [plugins/base](https://github.com/OutOfBedlam/tine/tree/main/plugins/base)

**Config**

```toml
[[flows.update]]
    ## Update name and value of fields and tags in a record with the new value and name.
    ## The toml syntax does not allow newlines within inline tables, 
    ## so all fields are specified in a single line
    set = [
        ## "field" is the field name to be updated.
        ## "name" is the new name of the field. If not specified, the original name is used.
        ## "value" is the new value of the field. If not specified, the original value is used.
        { field = "my_name", name = "new_name" },
        { field = "my_int", value = 10 },
        { field = "my_float", value = 9.87, name = "new_float" },
        { field = "flag", value = true, name = "new_flag" },
        { tag = "_in", value = "my" },
    ]
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

### SELECT

*Source* [plugins/base](https://github.com/OutOfBedlam/tine/tree/main/plugins/base)

**Config**

```toml
[[flows.select]]
    ## Selects fields in a record with the given field names.
    ## "*" means all fields.
    ## if item starts with "#", it specifies a tag name.
    ## '**' means all tags and fields which is default and equivalent to ["#*", "*"]
    includes = ["#_ts", "#_in", "*"]
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

### INJECT

*Source* [plugins/base](https://github.com/OutOfBedlam/tine/tree/main/plugins/base)

```toml
[[flows.inject]]
    id = "my_flow"
```

This flow creates an injection point for the pipeline, 
allowing applications to inject their own functions as flows with the specified ID.

**Example**

```toml
```

*Run*

```sh
```

*Output*

```json
```

### OLLAMA

*Source* [plugins/ollama](https://github.com/OutOfBedlam/tine/tree/main/plugins/ollama)

**Config**

```toml
## Ollama plugin configuration
## It takes variables from the input record
## The record should have the following keys:
## "prompt" string field
## "model" string field (e.g. "phi3")
## "stream" boolean field (default is false)
[[flows.ollama]]
    ## Ollama server address
    address = "http://127.0.0.1:11434"
    ## default model if the input record does not have the "model" field
    model = "phi3"
    ## timeout for waiting the response from the Ollama server
    timeout = "15s"
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