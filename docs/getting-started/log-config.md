# Log Config

Default log configuration.

```toml
[log]
    ## log file path, "-" for stdout, "" for no log
    ## default is ""
    path = ""

    ## log level: DEBUG, INFO, WARN, ERROR
    level = "INFO"

    ## max log file size in MB
    max_size = 100

    ## max log file age in days
    max_age = 7

    ## max log file backups
    max_backups = 10

    ## compress log backup files
    compress = false

    ## chown the log file (empty for current user)
    ## ignored on windows
    chown = ""

    ## use ansi color in log (useful for console output)
    no_color = false

    ## add source file and line to log
    add_source = false
```