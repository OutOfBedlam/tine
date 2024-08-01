# Inlets

### ARGS

> `import github.com/OutOfBedlam/tine/plugin/inlets/args`

Parse the command line arguments that following the `--` double dash. It assumes all arguments after the `--` are formed as key=value paris. And it pass them to the next step as a record.

For example, if the command line arguments are `tine some.toml -- --key1 value1 --key2 value2` then the record passed to the next step will be `{key1="value1", key2="value2"}`.

If value has prefix `base64+` followed by `http://`, `https://`, or `file://`, then the value is base64 encoded string of the content that are fetched from the URL or file.

If value has prefix `binary+` followed by `http://`, `https://`, or `file://`, then a `BinaryField` will be added instead of `StringField` within content that are fetched from the URL or file.

### EXEC

> `import github.com/OutOfBedlam/tine/plugin/inlets/exec`

Execute external command and yields records for the output of stdout of the command.&#x20;

#### Parameters

* commands (string array)
* environment (string array)
* prefix (string)
* timeout (string duration)
* ignore\_error (bool)
* interval (string duration)
* count (integer)
* trim\_space (bool)

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

