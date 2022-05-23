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

### Expand and Parse
`curl -X POST -d '{"query": "100 main st buffalo ny"}' <host>:8080/expandparser`

Original query is parsed and added with `"type": "query"`.  
All query expansions are parsed and added with `"tpye": "expansion"`

** Response **  

```
[
    {
        "data": "100 main st buffalo ny",
        "parsed": [
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
        ],
        "type": "query"
    },
    {
        "data": "100 main saint buffalo ny",
        "parsed": [
            {
                "label": "house_number",
                "value": "100"
            },
            {
                "label": "road",
                "value": "main"
            },
            {
                "label": "city",
                "value": "saint buffalo"
            },
            {
                "label": "state",
                "value": "ny"
            }
        ],
        "type": "expansion"
    },
    {
        "data": "100 main saint buffalo new york",
        "parsed": [
            {
                "label": "house_number",
                "value": "100"
            },
            {
                "label": "road",
                "value": "main"
            },
            {
                "label": "city",
                "value": "saint buffalo"
            },
            {
                "label": "state",
                "value": "new york"
            }
        ],
        "type": "expansion"
    },
    {
        "data": "100 main street buffalo ny",
        "parsed": [
            {
                "label": "house_number",
                "value": "100"
            },
            {
                "label": "road",
                "value": "main street"
            },
            {
                "label": "city",
                "value": "buffalo"
            },
            {
                "label": "state",
                "value": "ny"
            }
        ],
        "type": "expansion"
    },
    {
        "data": "100 main street buffalo new york",
        "parsed": [
            {
                "label": "house_number",
                "value": "100"
            },
            {
                "label": "road",
                "value": "main street"
            },
            {
                "label": "city",
                "value": "buffalo"
            },
            {
                "label": "state",
                "value": "new york"
            }
        ],
        "type": "expansion"
    }
]
```
