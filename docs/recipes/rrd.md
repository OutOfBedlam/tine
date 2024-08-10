# RRD

{% hint style="info" %}
rrd plugins requires Tine to be built with `-tags rrd` which need `librrd-dev` package to be installed in advance.
{% endhint %}

### Save data into RRD

**rrd_out.toml**

```toml
[log]
    path  = "-"
    level = "INFO"

[[inlets.load]]
    loads = [1, 5, 15]

[[outlets.rrd]]
    ## RRD database file path
    path = "./tmp/rrd_out.rrd"
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

**Execute**

```sh
tine run rrd_out.toml
```


### Read data from RRD and generate graph

```toml
[log]
    path  = "-"
    level = "INFO"

[[inlets.rrd_graph]]
    interval = "1s"
    range = "15m"
    title = "Test Graph"
    size = [450,130]
    units_length = 5
    v_label = "Load"
    theme = "gchart2"
    zoom = 1.5
    border = 0
    units_exponent = 0
    path = "./tmp/test.rrd"
    fields = [
        { ds = "load1", cf="AVERAGE", type="line", min="%3.1lf", max="%3.1lf", avg="%3.1lf", last="%3.1lf\\n" },
        { ds = "load5", cf="AVERAGE", type="line", min="%3.1lf", max="%3.1lf", avg="%3.1lf", last="%3.1lf\\n"},
        { ds = "load15", cf="AVERAGE", type="area", min="%3.1lf", max="%3.1lf", avg="%3.1lf", last="%3.1lf\\n"},
    ]

[[outlets.image]]
    path = "./tmp/rrd.png"
    overwrite = true

```