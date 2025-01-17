# Web Page Screenshot with Headless Chrome

{% hint style="warning" %}
The `chrome_snap` plugin relies on having Google Chrome browser installed beforehand.
{% endhint %}

The following recipe demonstrates how to capture two web pages and save them as image files using the `chrome_snap` plugin.

## Static Target Recipe

**example.toml**

```toml
#[log]
#    path = "-"
#    level = "debug"
[[inlets.file]]
    data = [
        '{"url":"https://tine.thingsme.xyz", "dst_path":"./chrome_snap_tine_docs.png"}', 
        '{"url":"https://github.com/OutOfBedlam/tine", "dst_path":"./chrome_snap_tine_github.png"}', 
    ]
    format = "json"
[[flows.chrome_snap]]
    url_field = "url"
    out_field = "snap"
    timeout = "15s"
[[outlets.image]]
    path_field = "dst_path"
    image_fields = ["snap"]
    overwrite = true
```

**Run**

```sh
$ ./tine run ./example.toml
```

**Output**

<figure><img src="./images/chrome_snap_tine_docs.png" alt="" width="563"><figcaption><p>chrome_snap_tine_docs.png</p></figcaption></figure>

<figure><img src="./images/chrome_snap_tine_github.png" alt="" width="563"><figcaption><p>chrome_snap_tine_github.png</p></figcaption></figure>

## User Input Recipe

**example.toml**

```toml
#!/path/to/tine run

#[log]
#    path = "-"
#    level = "info"
[[inlets.args]]
[[flows.chrome_snap]]
    url_field = "url"
    out_field = "snap"
    timeout = "15s"
[[outlets.image]]
    path_field = "out"
    image_fields = ["snap"]
    overwrite = true
```

**Run**

```sh
$ chmod +x example.toml

$ ./example.toml -- url=https://github.com out=./out.png

$ ls -l out.png
```