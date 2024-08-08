# OLLAMA

### Generates text via OLLAMA from user text by command line argument.

**ollama.toml**

```toml
[log]
    path  = "-"
    level = "INFO"

[[inlets.args]]

[[flows.ollama]]
    address = "http://127.0.0.1:11434" # <- OLLAMA API
    model = "phi3"                     # <- Model
    stream = true                      # <- Response mode, see the difference in the Result
    timeout = "120s"

[[outlets.file]]
    path = "-"
    format = "json"
```

**Execute**

```sh
tine run ollama.toml -- --prompt="hello, who are you?"
```

**Result**

- When set `stream=false`

{% code overflow="wrap" %}
```json
{"_ts":1722166588,"created_at":1722166588,"done":true,"done_reason":"stop", "response":" I'm Phi, an AI developed by Microsoft. How can I help you today?", "total_duration":1838575168}
```
{% endcode %}

- When set `stream=true`

{% code overflow="wrap" %}
```json
{"_ts":1722166487,"created_at":1722166483,"done":false,"done_reason":"","response":" Hello","total_duration":0}
{"_ts":1722166487,"created_at":1722166484,"done":false,"done_reason":"","response":"!","total_duration":0}
{"_ts":1722166487,"created_at":1722166484,"done":false,"done_reason":"","response":" I","total_duration":0}
{"_ts":1722166487,"created_at":1722166484,"done":false,"done_reason":"","response":"'","total_duration":0}
{"_ts":1722166487,"created_at":1722166484,"done":false,"done_reason":"","response":"m","total_duration":0}
{"_ts":1722166487,"created_at":1722166484,"done":false,"done_reason":"","response":" Ph","total_duration":0}
{"_ts":1722166487,"created_at":1722166484,"done":false,"done_reason":"","response":"i","total_duration":0}
{"_ts":1722166487,"created_at":1722166485,"done":false,"done_reason":"","response":",","total_duration":0}
{"_ts":1722166487,"created_at":1722166485,"done":false,"done_reason":"","response":" an","total_duration":0}
{"_ts":1722166487,"created_at":1722166485,"done":false,"done_reason":"","response":" A","total_duration":0}
{"_ts":1722166487,"created_at":1722166485,"done":false,"done_reason":"","response":"I","total_duration":0}
{"_ts":1722166487,"created_at":1722166485,"done":false,"done_reason":"","response":" developed","total_duration":0}
{"_ts":1722166487,"created_at":1722166485,"done":false,"done_reason":"","response":" by","total_duration":0}
{"_ts":1722166487,"created_at":1722166485,"done":false,"done_reason":"","response":" Microsoft","total_duration":0}
{"_ts":1722166487,"created_at":1722166485,"done":false,"done_reason":"","response":" to","total_duration":0}
{"_ts":1722166487,"created_at":1722166485,"done":false,"done_reason":"","response":" help","total_duration":0}
{"_ts":1722166487,"created_at":1722166486,"done":false,"done_reason":"","response":" answer","total_duration":0}
{"_ts":1722166487,"created_at":1722166486,"done":false,"done_reason":"","response":" questions","total_duration":0}
{"_ts":1722166487,"created_at":1722166486,"done":false,"done_reason":"","response":" and","total_duration":0}
{"_ts":1722166487,"created_at":1722166486,"done":false,"done_reason":"","response":" assist","total_duration":0}
{"_ts":1722166487,"created_at":1722166486,"done":false,"done_reason":"","response":" with","total_duration":0}
{"_ts":1722166487,"created_at":1722166486,"done":false,"done_reason":"","response":" tasks","total_duration":0}
{"_ts":1722166487,"created_at":1722166486,"done":false,"done_reason":"","response":".","total_duration":0}
{"_ts":1722166487,"created_at":1722166486,"done":false,"done_reason":"","response":" How","total_duration":0}
{"_ts":1722166487,"created_at":1722166486,"done":false,"done_reason":"","response":" can","total_duration":0}
{"_ts":1722166487,"created_at":1722166486,"done":false,"done_reason":"","response":" I","total_duration":0}
{"_ts":1722166487,"created_at":1722166486,"done":false,"done_reason":"","response":" assist","total_duration":0}
{"_ts":1722166487,"created_at":1722166486,"done":false,"done_reason":"","response":" you","total_duration":0}
{"_ts":1722166487,"created_at":1722166487,"done":false,"done_reason":"","response":" today","total_duration":0}
{"_ts":1722166487,"created_at":1722166487,"done":false,"done_reason":"","response":"?","total_duration":0}
{"_ts":1722166487,"created_at":1722166487,"done":true,"done_reason":"stop","response":"","total_duration":5411992794}
```
{% endcode %}
