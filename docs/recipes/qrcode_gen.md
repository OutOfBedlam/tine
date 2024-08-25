# QRCode Generator

This recipe demonstrates how to generate QR Code.

<figure><img src="./images/qrcode_gen_output.png" alt="" width="200"><figcaption><p>QRCode output</p></figcaption></figure>

**example.toml**

```toml
[[inlets.args]]
[[flows.qrcode]]
    input_field = "in"
    output_field = "qrcode"
    # QRCode width should be < 256
    width = 11
    # background_transparent = true
    background_color = "#ffffff"
    foreground_color = "#000000"
    # logo image should only has 1/5 width of QRCode at most (.png or .jpeg)
    logo = "./plugins/qrcode/testdata/tine_x64.png"
[[outlets.image]]
    path_field = "out"
    image_fields = ["qrcode"]
    overwrite = true
```

**Run**

```sh
tine run ./example.toml -- in="https://tine.thingsme.xyz" out=./output.png
```
