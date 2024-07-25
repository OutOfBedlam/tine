# TINE Examples

## How to use TINE as a library in an applicaton.

- [helloworld](./helloworld/helloworld.go)

## How to set a custom outlet.

- [custom_out](./custom_out/custom_out.go)

## How to set a custom inlet.

- [custom_in](./custom_in/custom_in.go)

## How to set a custom flow.

- [custom_flow](./custom_flow/custom_flow.go)

## How to use piplelines as a HTTP handler

- [httpsvr](./httpsvr/httpsvr.go)


## How to collect metrics, save it into RRD and display rrdgraph on web page

- [rrd_graph_web](./rrd_graph_web/rrd_graph_web.go)

<!-- ![image](./rrd_graph_web/rrd_graph_web.png) -->
<img src="./rrd_graph_web/rrd_graph_web.png" alt="image" width="300" height="auto">

## How to collect metrics into Sqlite and display it on web page

This example also shows how to utilize HTTP query parameters 
as variables of Go Templates to build pipeline configuration.

- [sqlite_graph_web](./sqlite_graph_web/sqlite_graph_web.go)