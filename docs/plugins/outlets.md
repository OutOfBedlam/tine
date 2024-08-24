# Outlets

### EXCEL

*Source* [plugins/excel](https://github.com/OutOfBedlam/tine/tree/main/plugins/excel)

**Config**

```toml
[[outlet.excel]]
    ## Save the records into Microsoft Excel file format.
    ## File path (*.xlsx) to save the records
    path = "./output.xlsx"
    ## It stores all records in memory first then write to the file 
    ## when the count of the buffered records in memory reaches to "records_per_file".
    ## So, every excel file it writes has "records_per_file" rows. (+ 1 header row)
    ## If records_per_file is 0, it will keep all records in memory (consuming memory) 
    ## and write to the file when the pipeline is closed.
    ## (default 20000)
    records_per_file = 20000
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

### FILE

*Source* [plugins/base](https://github.com/OutOfBedlam/tine/tree/main/plugins/base)

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

### HTTP

*Source* [plugins/http](https://github.com/OutOfBedlam/tine/tree/main/plugins/http)

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

### IMAGE

*Source* [plugins/image](https://github.com/OutOfBedlam/tine/tree/main/plugins/image)

**Config**

```toml
[[outlets.image]]
    ##
    # overwrite = true
    #
    ## The path to the output file.
    ## If the file exists and overwrite is false, it will be append "_"+field_name+"_"+sequence number before the extension.
    ## If the file exists and overwrite is true, it will be append "_"+field_name before the extension.
    ## supported formats: png, jpeg, gif, bmp
    path = "./output.png"
    ## fields that contains the image data
    ## if not specified, it will find all binary fields that has Content-type "image/*"
    # image_fields = ["image"]
    ## Quality of jpeg image (jpeg only)
    ## 1 ~ 100
    # jpeg_quality = 75
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

### INFLUX

*Source* [plugins/influx](https://github.com/OutOfBedlam/tine/tree/main/plugins/influx)

**Config**

```toml
[[outlets.influx]]
    ## The name of the database to write to
    ##
    ## If the database does not exist, create first by running:
    ## curl -XPOST 'http://localhost:8086/query' --data-urlencode 'q=CREATE DATABASE "metrics"'
    ##
    db = "metrics"
    ## The URL of the InfluxDB server
    ## Or file path to write to a file
    # path = "-"
    path = "http://127.0.0.1:8086/write?db=metrics"
    ## The tags to add to the metrics
    ## If 'value' is not set, the value of the tag will be taken from the record
    tags = [
        {name="dc", value="us-east-1"},
        {name="env", value="prod"},
        {name="_in"}
    ]

    ## Write timeout, especially for the HTTP request
    timeout = "3s"
    ## Debug mode for logging the response message from the InfluxDB 
    debug = true
```

**Example**

```toml
[log]
    path = "-"
    level = "info"
[defaults]
    interval = "5s"
[[inlets.load]]
    loads = [1,5,15]
[[inlets.mem]]
[[inlets.host]]
[[flows.merge]]
    wait_limit = "5s"
    name_infix = "."
[[outlets.influx]]
    db = "metrics"
    path = "http://127.0.0.1:8086/write?db=metrics"
    tags = [
        {name="dc", value="us-east-1"},
        {name="env", value="prod"},
        {name="_in"}
    ]
    timeout = "3s"
    debug = false
```

### MQTT

*Source* [plugins/mqtt](https://github.com/OutOfBedlam/tine/tree/main/plugins/mqtt)

**Config**

```toml
[[outlets.mqtt]]
    ## mqtt server address
    server = "127.0.0.1:1883"
    ## mqtt topic to publish
    topic  = "topic_to_publish"
    ## publish QoS, supports 0, 1
    qos = 1
    ## timeout for CONN and PUBLISH (default: 3s)
    timeout = "3s"
    ## output format
    format = "csv"
    ## output fields
    fields = []
    ## output compression
    compress = ""
    ## time format (default: s)
    ##  s, ms, us, ns, Golang timeformat string")
    ##  e.g. timeformat = "2006-01-02 15:04:05 07:00"
    timeformat = "s"
    ## timezone (default: Local)
    ##  e.g. tz = "Local"
    ##  e.g. tz = "UTC"
    ##  e.g. tz = "America/New_York"
    tz = "Local"
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
[[outlets.telegram]]
    token = "<bot_token>"
      ## If the input record has "chat_id" INT field it will be used,
    ## otherwise the default chat_id will be used
    ## If the input record doesn't have "chat_id" field,
    ## and the default "chat_id" is not set, the record will be ignored
    chat_id = "<chat_id>"
    debug = false
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

### TEMPLATE

*Source* [plugins/template](https://github.com/OutOfBedlam/tine/tree/main/plugins/template)

**Config**

```toml
[[outlets.template]]
    ## File path to write the output, "-" means stdout, "" means discard
    path = "-"
    
    ## Overwrite the file if "path" is a file and it already exists
    overwrite = false

    ## Output the data in column mode
    column_series = "json"

    ## Templates in string
    templates = [
        """{{ range . }}TS: {{ (index ._ts).Format "2006 Jan 02 15:04:05" }} INLET:{{ index ._in }} load1: {{ index .load1 }} {{ end }}\n"""
    ]
    ## Template files to load
    templateFiles = []

    ## Timezone to use for time formatting
    timeformat = "s"

    ## Timezone to use for time formatting
    tz = "Local"

    ## Decimal places for float values
    decimal = -1
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
