
This plugin requires rrdtool installed in advance.

https://github.com/oetiker/rrdtool-1.x


Ubuntu:

```
sudo apt install librrd-dev
```

macOS

```
brew install rrdtool
```

Build with `-tags rrd`

```
go build -tags rrd
```


Usage example

```toml
[[outlets.rrd]]
    path       = "/data/test.rrd"
    fields     = ["load1", "load5"]
    time_field = "_ts"
    step = "1s"
```
