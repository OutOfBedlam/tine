# Syslog Receiver

### Receive rsyslog messages

**syslog.toml**

```toml
[[inlets.syslog]]
    ## Listen address
    ## e.g. tcp://:5514, udp://:5514, unix:///var/run/syslog.sock
    address = "udp://:5516"

[[outlets.file]]
    path = "-"
    format = "json"
```

**Output**

```json
{
    "appname":"login",
    "facility_code":0,
    "hostname":"local.local",
    "message":"USER_PROCESS: 17309 ttys004",
    "procid":"17309",
    "remote_host":"127.0.0.1",
    "severity_code":5,
    "timestamp":1724490558
}
{
    "appname":"sudo",
    "facility_code":1,
    "hostname":"local.local",
    "message":"getgrouplist_2 called triggering group enumeration",
    "procid":"17314",
    "remote_host":"127.0.0.1",
    "severity_code":5,
    "timestamp":1724490558
}
```
