# Libpostal REST

## API Example

Replace <host> with your host

### Build

Go 1.13+. Build steps for older go versions may vary.

```
go build .
ls ./libpostal-rest
```

### Parser
`curl -X POST -d '{"query": "100 main st buffalo ny"}' <host>:8080/parser`

** Response **
```
[
  {
    "label": "house_number",
    "value": "100"
  },
  {
    "label": "road",
    "value": "main st"
  },
  {
    "label": "city",
    "value": "buffalo"
  },
  {
    "label": "state",
    "value": "ny"
  }
]
```

### Expand
`curl -X POST -d '{"query": "100 main st buffalo ny"}' <host>:8080/expand`

** Response **
```
[
  "100 main saint buffalo new york",
  "100 main saint buffalo ny",
  "100 main street buffalo new york",
  "100 main street buffalo ny"
]
```
