# RRD Graph

> [!IMPORTANT]
> rrd plugins requires TINE to be built with `-tags rrd` which need `librrd-dev` package to be installed in advance.

## Read data from RRD and generate graph

```toml
[log]
    filename = "-"
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