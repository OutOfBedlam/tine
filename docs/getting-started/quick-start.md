# Quick start

Define a simple pipeline description in a TOML file.

Copy the example below and save it as `cpu.toml`.

```toml
[[inlets.cpu]]
    interval = "3s"
[[flows.select]]
    includes = ["**"]
[[outlets.file]]
    path  = "-"
```

Run tine passing the `cpu.toml` path  as argument.

```bash
tine run ./cpu.toml
```

Press ^C to stop.

{% code lineNumbers="true" %}
```
2024-07-28 21:53:50 INF pipeline cpu.toml start inlets=1 flows=1 outlets=1
cpu,1722171230,8.305903745046313
cpu,1722171233,11.401743796163146
cpu,1722171236,9.507754551438795
cpu,1722171239,8.965748824811843
^C
2024-07-28 21:54:01 INF pipeline cpu.toml stop
```
{% endcode %}

Line 1 and 7 are log messages that are printed out on stdout by default with an INF (info) level.

From line 2 to 5, the TINE process prints out a CSV line every 3 seconds until it is stopped by pressing ^C.

Each line in the content represents the "cpu" inlet, which generates data consisting of a timestamp in UNIX epoch time and the corresponding system CPU usage percentage.

Add `format="json"` at the end of the file, it changes the out `outlets.file` applying JSON format instead of CSV which is default.

```toml
[[outlets.file]]
    path  = "-"
    format = "json"
```

```
2024-07-28 22:04:16 +0900 INF pipeline example.toml start inlets=1 flows=1 outlets=1
[{"_in":"cpu","_ts":1722171856,"total_percent":8.296722122944592}]
[{"_in":"cpu","_ts":1722171859,"total_percent":5.32841823082893}]
[{"_in":"cpu","_ts":1722171862,"total_percent":5.522088353490716}]
[{"_in":"cpu","_ts":1722171865,"total_percent":4.525645323428138}]
^C
2024-07-28 22:04:26 +0900 INF pipeline example.toml stop
```

### Using shebang

Edit the `cpu.toml` file and add the following line at the beginning: `#!/path/to/tine run`. Replace `/path/to/tine` with the actual path to the tine executable file on your system.

The `chmod +x cpu.toml` to make it executable.

Run

```bash
$./cpu.toml
2024-07-28 22:12:13 +0900 INF pipeline example.toml start inlets=1 flows=1 outlets=1
[{"_in":"cpu","_ts":1722172333,"total_percent":8.287984091499398}]
[{"_in":"cpu","_ts":1722172336,"total_percent":6.224899598711208}]
[{"_in":"cpu","_ts":1722172339,"total_percent":5.395442359208408}]
^C
2024-07-28 22:12:19 +0900 INF pipeline example.toml stop
```

