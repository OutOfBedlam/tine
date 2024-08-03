# Extras

### RRD

*Source* [x/rrd](https://github.com/OutOfBedlam/tine/tree/main/x/rrd)

**Config**

```toml
[[outlets.rrd]]
    ## RRD database file path
    path = "./tmp/test.rrd"
    ## Overwrite the file if it already exists
    overwrite = false
    ## Input data interval (minimum is 1s)
    step = "1s"
    ## default heartbeat that will be used for all data sources
    heartbeat = "2s"
    ## which time field to use for the data source (default is "_ts")
    time_field = "_ts"
    ## field field to map to the data source
    ## ds    Data Source Name (if not specified, field name will be used)
    ##       If ds contains invalid characters for RRD(e.g. ':'), it will be replaced with "_"
    ## dst   Data Source Type
    ##       GAUGE, COUNTER, DCOUNTER, DERIVE, DDERIVE, ABSOLUTE, COMPUTE
    ## heartbeat should be larger than step
    ## min   minimum value, "U" means unknown,
    ##       if input value is less than min, it will be treated as unknown
    ## max   maximum value, "U" means unknown,
    ##       if input value is greater than max, it will be treated as unknown
    ## rpn   Reverse Polish Notation expression
    fields = [
        { field="load1",  dst="GAUGE", heartbeat="2s", min=0.0, max="U", rpn="" },
        { field="load5",  dst="GAUGE", heartbeat="2s", min=0.0, max="U", rpn="" },
        { field="load15", dst="GAUGE", heartbeat="2s", min=0.0, max="U", rpn="" },
    ]

    ## Round Robin Archive
    ##
    ## cf    consolidation function
    ##       AVERAGE, MIN, MAX, LAST
    ## xxf   xfiles factor, how long to consider data as known when unknown data comes in.
    ##       It's a value between 0 and 1, with the default being 0.5.
    ##       For example, if xff is set to 0.5 and 50% of the data points are known (1, unknown, unknown, 1),
    ##       the average will be 1. If 75% of the data points are unknown (1, unknown, unknown, unknown),
    ##       the result will be unknown.
    ## steps how many steps to use for calculation of this RRA
    ##       1 means every data point is used which is equal to the 'LAST' function
    ## rows  how many rows to store in this RRA
    rra = [
        { cf = "AVERAGE", steps = "1s", rows="3h" },
        { cf = "AVERAGE", steps = "1m", rows="3d" },
        { cf = "AVERAGE", steps = "1h", rows="30d" },
        { cf = "AVERAGE", steps = "1d", rows="13M" },
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

### RRD-GRAPH

*Source* [x/rrd](https://github.com/OutOfBedlam/tine/tree/main/x/rrd)

**Config**

```toml
[[inlets.rrd_graph]]
    ## RRD database file path to read
    path = "./tmp/test.rrd"
    ## rrdtool graph generation interval
    ## if not set, it will generate graph once
    interval = "1s"
    ## graph title
    title = "Test Graph"
    ## time range: from now - range to now
    range = "15m"
    ## graph size [width, height]
    size = [450,130]
    ## border
    border = 0
    ## units length
    units_length = 5
    ## units exponent
    units_exponent = 0
    ## watermark text (default is pipeline name)
    watermark = "My Watermark"
    ## vertical label
    v_label = "Load"
    ## zoom factor
    zoom = 1.5
    ## graph color theme
    theme = "gchart2"
    ## specify each color override for theme colors
    # back = "#ffffff"
    # canvas = "#ffffff"
    # font = "#000000"
    ## data sources to draw
    ## ds     data source name in rrd
    ## cf     consolidation function (AVERAGE, MIN, MAX, LAST)
    ## type   line, area, stack, etc.
    ## color  color code (e.g. "#ff0000") override for theme colors
    ## name   data source name format (e.g. name="%-6s\\n")
    ## min    minimum value format (e.g. min="%3.1lf"), omit not to show
    ## max    maximum value format (e.g. max="%3.1lf"), omit not to show
    ## avg    average value format (e.g. avg="%3.1lf"), omit not to show
    ## last   last value format (e.g. last="%3.1lf"), omit not to show
    fields = [
        { ds = "load1", cf="AVERAGE", type="line", name="%-6s", min="%3.1lf", max="%3.1lf", avg="%3.1lf", last="%3.1lf\\n"},
        { ds = "load5", cf="AVERAGE", type="line", name="%-6s", min="%3.1lf", max="%3.1lf", avg="%3.1lf", last="%3.1lf\\n"},
        { ds = "load15", cf="AVERAGE", type="area", name="%-6s", min="%3.1lf", max="%3.1lf", avg="%3.1lf", last="%3.1lf\\n"},
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
