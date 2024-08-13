# Outlets

### EXCEL

*Source* [plugin/outlets/excel](https://github.com/OutOfBedlam/tine/tree/main/plugin/outlets/excel)

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

*Source* [plugin/outlets/file](https://github.com/OutOfBedlam/tine/tree/main/plugin/outlets/file)

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

*Source* [plugin/outlets/http](https://github.com/OutOfBedlam/tine/tree/main/plugin/outlets/http)

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

*Source* [plugin/outlets/image](https://github.com/OutOfBedlam/tine/tree/main/plugin/outlets/image)

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

### MQTT

*Source* [plugin/outlets/mqtt](https://github.com/OutOfBedlam/tine/tree/main/plugin/outlets/mqtt)

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

*Source* [plugin/outlets/telegram](https://github.com/OutOfBedlam/tine/tree/main/plugin/outlets/telegram)

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

*Source* [plugin/outlets/template](https://github.com/OutOfBedlam/tine/tree/main/plugin/outlets/template)

**Config**

```toml
[[outlets.template]]
    ## File path to write the output, "-" means stdout, "" means discard
    path = "-"
    ## Overwrite the file if "path" is a file and it already exists
    overwrite = false
    ## Templates in string
    templates = [
        """{{ range . }}TS: {{ (index ._ts).Format "2006 Jan 02 15:04:05" }} INLET:{{ index ._in }} load1: {{ index .load1 }} {{ end }}\n"""
    ]
    ## Template files to load
    templateFiles = []
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
