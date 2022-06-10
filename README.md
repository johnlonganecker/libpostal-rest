# Libpostal REST

## Build

Go 1.13+. Build steps for older go versions may vary.

```
go build .
ls ./libpostal-rest
```

## Running

### Environment Variables

This service can be configured with the following environment variables:

`LISTEN_HOST` - hostname or IP address to listen on - default: 0.0.0.0 (all interfaces)

`LISTEN_PORT` - port the server is listening on - default: 8080

`LOG_LEVEL` - setting the global log level - default `info` (more on that further below)

`LOG_STRUCTURED` - if set to `true`, the logger generates structured JSON log output, otherwise it will create "pretty" output - default: `false`

`PROMETHEUS_ENABLED` - set to `true` enables Prometheus metrics collector and endpoint, 'false' or missing disables it - default: `false`

`PROMETHEUS_PORT` - port number where Prometheus exposes the `/metrics` endpoint - default: 9090

### Logging

Logging has been implemented using [zerolog](https://github.com/rs/zerolog) which is a super-fast structured logger outputting JSON. However, `zerolog` also allows outputting non-structured ("pretty printed") log lines, which is the default setting. `zerolog` also supports different log levels which are as follows:

* `panic`
* `fatal`
* `error`
* `warn`
* `info`
* `debug`
* `trace`

For example, to change the default log level from `info` to `debug` one would set the following environment variable:
```
export LOG_LEVEL='debug'
```

> Please note, if the string set in the environment variable is not one of the valid ones listed above, the service will log a warning and used the default level `info`.

### Metrics Generation

The service has been instrumented with the [Prometheus client library](https://github.com/prometheus/client_golang). By default metrics collection and the Prometheus endpoint is disabled. It can be enabled by setting the environment variable `PROMETHEUS_ENABLED` to `true`. Once enabled, the metrics can be scraped at the following endpoint:
```
http://<LISTEN_HOST>:<PROMETHEUS_PORT>/metrics
```

The metrics endpoint exposes the default [golang metrics](https://github.com/prometheus/client_golang/blob/v1.12.2/prometheus/collectors/go_collector_latest.go#L25) as well as the following custom metrics of the service:
* `libpostal_expand_reqs_total` - The total number of processed expand requests
* `libpostal_expand_durations_historgram_ms` - Latency distributions of expand calls in milliseconds
* `libpostal_parse_reqs_total` - The total number of processed parse requests
* `libpostal_parse_durations_historgram_ms` - Latency distributions of parse calls in milliseconds
* `libpostal_expandparse_reqs_total` - The total number of processed expandparse requests
* `libpostal_expandparse_durations_historgram_ms` - Latency distributions of expandparse calls in milliseconds

## API Example

Replace `<host>` with your host, e.g. `localhost`

### Parser
**Request**
```
curl -X POST -d '{"query": "100 main st buffalo ny"}' http://<host>:8080/parser
```

**Response**
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

### Expand without language options
**Request**

```
curl -X POST -d '{"query": "100 main st buffalo ny"}' http://<host>:8080/expand
```

**Response**

```
[
  "100 main saint buffalo new york",
  "100 main saint buffalo ny",
  "100 main street buffalo new york",
  "100 main street buffalo ny"
]
```

### Expand **with** language options
**Request**

> **IMPORTANT NOTE:** if the `langs` array contains an invalid language specifier, e.g. `"xyz"`, the string will **not** be expanded for that specified language input and instead the original query string will be returned.

```
curl -X POST -d "{\"query\": \"100 main st buffalo ny\", \"langs\": [\"de\"]}" http://localhost:8080/expand
```

**Response**

```
[
  ""100 main sankt buffalo ny""
]
```

Incorrect language specified

**Request**
```
curl -X POST -d "{\"query\": \"100 main st buffalo ny\", \"langs\": [\"de\", \"xyz\"]}" http://localhost:8080/expand
```

**Response**

```
[
    "100 main sankt buffalo ny",
    "100 main st buffalo ny"
]
```

### Expand and Parse without language option
**Request**

```
curl -X POST -d '{"query": "100 main st buffalo ny"}' http://<host>:8080/expandparser
```

Original query is parsed and added with `"type": "query"`.
All query expansions are parsed and added with `"type": "expansion"`

**Response**

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

### Expand and Parse **with** language option
**Request**
```
curl -X POST -d "{\"query\": \"100 main st buffalo ny\", \"langs\": [\"fr\"]}" http://localhost:8080/expandparser
```

**Response**

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
    }
]
```