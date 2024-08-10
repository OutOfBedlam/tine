# Concept

## Entities

### Pipeline

A pipeline consists of multiple inlets, flows and outlets.

A pipeline is a series of steps or processes that are used to extract, transform, and load (ETL) data from one or more sources to one or more destinations. It is a common concept in data engineering and is used to automate the movement and transformation of data.

Here's an example of how a data pipeline works:

**Inlets**: The first step in a data pipeline is to extract or retrieve data from various sources such as databases, APIs, files, or streaming platforms. This can involve querying databases, making API requests, or reading files by using the appropriate inlets.

**Flows**: Once the data is retrieved, it often needs to be transformed or cleaned to make it suitable for analysis or destinations. This can involve tasks like filtering, aggregating, joining, or applying calculations to the data. Data transformation can be done using various flows. In some cases, additional processing or analysis may be required on the transformed data. This can involve running machine learning algorithms, performing statistical analysis, or generating reports.

**Outlets**: After the data is transformed, it needs to be exported/passed to a destination such as external system, a database, or a file system. This step involves writing the transformed data to the destination in a structured format.

```toml
[[inlets.cpu]]
    interval = "3s"
[[flows.select]]
    includes = ["**"]
[[outlets.file]]
```

<figure><img src="../.gitbook/assets/pipeline-cpu-select-file.png" alt="" width="563"><figcaption><p>A simple pipeline</p></figcaption></figure>

The example above is extracting data of system's CPU usage by `inlets.cpu` and pass it to `flows.select` which is simply pass to the only outlet `outlets.file` that configured to print out on stdout.

### Record

The data flowing through a pipeline consists of a sequence of records. Each record contains one or more fields and may have zero or more tags. For instance, in the given pipeline definition, the `inlets.load` component retrieves the average system load for 1 minute, 5 minutes, and 15 minutes, and passes this information as a record every 3 seconds to the next component.

```toml
[[inlets.load]]
    interval = "3s"
    loads = [1, 5, 15]
[[outlets.file]]
    decimal = 2
```

The result of this pipeline might be similar to the following.

```csv
1.46,1.43,1.61
1.55,1.45,1.62
1.51,1.44,1.61
...
```

The first line of the result can be visualized in the following diagram.

<figure><img src="../.gitbook/assets/pipeline-load-file.png" alt="" width="563"><figcaption><p>A record</p></figcaption></figure>

The first record has three fields named `load1`, `load5`, and `load15`. The two tags `_in` that informs which inlet makes the data, and `_ts` (timestamp) which shows when the data is captured, are not exported into the result because tags are hidden from the final result by default.

To include tags in the result, you can use the `flows.select` expression with `includes = ["**"]`. This expression selects which fields can be passed to the next step and converts specified tags, indicated by `#tag_name`, into fields. In the example below, the `**` notation is used to include all tags and fields.

```toml
[[inlets.load]]
    interval = "3s"
    loads = [1, 5, 15]
[[flows.select]]
    includes = ["**"]
[[outlets.file]]
    path  = "-"
    decimal = 2
```

The pipeline generates the result might be like below.

```csv
load,1722427244,1.87,1.93,1.83
load,1722427247,1.87,1.93,1.83
load,1722427250,1.80,1.91,1.83
```

The first column "load" came from the value of `_in` tag followed by `_ts` tag's value and all other fields in order.

To make the result clearer, change the output format to "json".

```toml
[[inlets.load]]
    interval = "3s"
    loads = [1, 5, 15]
[[flows.select]]
    includes = ["**"]
[[outlets.file]]
    format = "json"
    decimal = 2
```

{% code overflow="wrap" %}
```json
{"_in":"load","_ts":1722428243,"load1":1.65,"load15":1.62,"load5":1.65}
{"_in":"load","_ts":1722428246,"load1":1.60,"load15":1.61,"load5":1.64}
{"_in":"load","_ts":1722428249,"load1":1.60,"load15":1.61,"load5":1.64}
```
{% endcode %}

### Field

A field is a name and value pair and can have zero or more tags. The tags of a field are independent of the record-level tags.

### Tag

A record and a field can have zero or more tags.

Tag is name and value pair which is similar to fields but tags are auxiliary data.

### Value

A "value" represents the unit of an actual value. The field and tag have its own value. Each value have specific type.

The available value types:

* BOOL boolean type (true or false)
* INT 64bits signed integer
* UINT 64bits unsigned integer
* FLOAT 64bits floating point
* STRING UTF-8 text
* TIME date and time in nano seconds
* BINARY a chunk of bytes

## Multiple Inlets and Outlets

A pipeline may have multiple inlets and outlets.

<figure><img src="../.gitbook/assets/pipeline-multi-inout.png" alt="" width="563"><figcaption><p>A pipeline with multiple inlets and outlets</p></figcaption></figure>

The example below shows a simple example.

```toml
[[inlets.cpu]]
    percpu = false
    interval = "1s"
    count = 3
[[inlets.load]]
    loads = [1, 5]
    interval = "1s"
    count = 2
[[outlets.file]]
    path = "-"
    format = "json"
    decimal = 2
[[outlets.file]]
    path = "-"
    format = "csv"
    decimal = 2
```

If you run this pipeline, it will print out the cpu usage and load average in CSV and JSON format.

```sh
$ tine run example.toml

2.19,3.45
{"load1":2.19,"load5":3.45}
{"total_percent":8.16}
8.16
{"load1":2.19,"load5":3.45}
{"total_percent":4.73}
2.19,3.45
4.73
2.62
{"total_percent":2.62}
```

### Merge records from the multiple inlets.

Generally we want to see the cpu usage and load average in a record by the observed *same* time.
The example below shows each inlet just yields its records and outlet prints them out in order.

```toml
[[inlets.cpu]]
    percpu = false
    interval = "1s"
    count = 3
[[inlets.load]]
    loads = [1, 5]
    interval = "1s"
    count = 2
[[outlets.file]]
    path = "-"
    format = "json"
    decimal = 2
```

```json
{"load1":2.14,"load5":2.11}
{"total_percent":8.16}
{"total_percent":11.97}
{"load1":2.05,"load5":2.09}
{"total_percent":12.26}
```

Let's add `[[flows.merge]]` to combine multiple records into a record. It yields new record that *merged* from multiple records that has *same* `_ts` tag.

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
<figure><img src="../.gitbook/assets/pipeline-merge.png" alt="" width="563"><figcaption><p>A pipeline with multiple inlets and merge</p></figcaption></figure>


`[[flows.merge]]` promotes the `_ts` tag to a field that has same name `_ts` and renames all other fields with the origin `_in` tag value followed by a dot and the original name. This operation is conceptually similar to joining tables in a relational database management system (RDBMS). It can be thought of as performing a `SELECT _ts, cpu.total_usage, load.load1, load.load5 FROM cpu, load WHERE cpu._ts = load._ts` query.

```json
{"_ts":1723255081,"cpu.total_percent":8.16,"load.load1":1.80,"load.load5":2.00}
{"_ts":1723255082,"cpu.total_percent":17.14,"load.load1":1.81,"load.load5":2.00}
{"_ts":1723255083,"cpu.total_percent":13.33}
```