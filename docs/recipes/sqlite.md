# SQLite


## Save data into SQLite

**sqlite_out.toml**

```toml
[defaults]
    interval = "5s"

[log]
    filename = "-"
    level = "DEBUG"

[[inlets.load]]
    loads = [1, 5, 15]

[[flows.flatten]]

[[outlets.sqlite]]
    path = "file::memory:?mode=memory&cache=shared"
    inits = [
        """
            CREATE TABLE IF NOT EXISTS metrics (
                time INTEGER,
                name TEXT,
                value REAL,
                UNIQUE(time, name)
            )
        """,
    ]
    actions = [
        [
            """ INSERT INTO metrics (time, name, value)
                VALUES (?, ?, ?)
            """, 
            "_ts", "name", "value"
        ],
    ]
```

**Execute**

```sh
tine run sqlite_out.toml
```

