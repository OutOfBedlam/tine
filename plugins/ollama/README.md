

Run Ollama docker

```sh
$ mkdir ./tmp/ollama

$ docker run -d -v ./tmp/ollama:/root/.ollama -p 11434:11434 --name ollama ollama/ollama
```

Pull model

```sh

curl http://localhost:11434/api/pull -d '{
  "name": "phi3"
}'

```

```sh
curl http://localhost:11434/api/generate  -d '{
  "model": "phi3",
  "prompt": "Why is the sky blue?",
  "format": "json",
  "stream": false
}'
```