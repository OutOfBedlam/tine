# Flows

### DAMPER

*Source* [plugin/flows/base](https://github.com/OutOfBedlam/tine/tree/main/plugin/flows/base)

**Config**

```toml
[[flows.damper]]
    ## damper makes the stream of records to be delayed by the given duration.
    ## It collects records those _ts time is older thant now-"interval" time, and send thme to the next flow in a time.
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

*Source* [plugin/flows/base](https://github.com/OutOfBedlam/tine/tree/main/plugin/flows/base)

**Config**

```toml
[[flows.dump]]
    ## dump writes the record to the log with the given log level.
    ## DEBUG | INFO | WARN | ERROR  (default: DEBUG)
	level = "DEBUG"
    ## The decimal format for float fields. (default: -1 which means no rounding)
    precision = 2
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

### MERGE

*Source* [plugin/flows/base](https://github.com/OutOfBedlam/tine/tree/main/plugin/flows/base)

**Config**

```toml
[[flows.merge]]
    wait_limit = "2s"
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

### UPDATE

*Source* [plugin/flows/base](https://github.com/OutOfBedlam/tine/tree/main/plugin/flows/base)

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

*Source* [plugin/flows/base](https://github.com/OutOfBedlam/tine/tree/main/plugin/flows/base)

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

### OLLAMA

*Source* [plugin/flows/ollama](https://github.com/OutOfBedlam/tine/tree/main/plugin/flows/ollama)

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