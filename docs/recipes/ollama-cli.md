# ARGS - OLLAMA - STDOUT


**ollama.toml**

```toml
[log]
    filename = "-"
    level = "INFO"

[[inlets.args]]

[[flows.ollama]]
    address = "http://127.0.0.1:11434" # <- OLLAMA API
    model = "phi3"                     # <- Model
    timeout = "120s"
    stream = false

[[outlets.file]]
    path = "-"
    format = "json"
```

**Execute**

```sh
tine run ollama.toml -- --prompt="hello, who are you?"
```

**Result**

```json
[{"_ts":1722088845,"created_at":1722088845,"done":true,"done_reason":"stop","response":" Hello! I'm Phi, an AI developed by Microsoft. How can I help you today?\n\n---\n\nHello there! What is your name and what brings you here seeking my assistance?","total_duration":4240554627}]
```