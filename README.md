# Jfrog Xray MS Teams Connector

## Motivation

This was created as teams does not accept the JSON payload from xray.

What this does is takes the Xray payload and convert it to a teams card format and forwards it onto teams.

## Usage

**Docker**

```bash
docker run -d [-e HTTPS_PROXY=""] squ1d123/xray-teams-connector:latest
```

**Native**

```bash
go build -tags netgo
./xray-teams-connector
```

**Configuring Xray**

When configuring Xray be sure to specify the header `Content-Type: application/json`.

The way the app works is by forwarding the request parameter `key=<webhookUrl>`
> The `webHookUrl` should not include `https://outlook.office.com/webhook/` as this is automaticall prefixed by the app

## Local Dev

To test locally just run binary and submit payload with url.

```bash
curl -L -d @test/payload.json -H "Content-Type: application/json" 'localhost:8000/webhook?key=<webhookUrl>'
```

## Monitoring

There is basic prometheus monitoring available on `localhost:8000/metrics`

```bash
curl  -s localhost:8000/metrics | egrep -i 'http_(request|response).*'
# HELP http_request_duration_microseconds The HTTP request latencies in microseconds.
# TYPE http_request_duration_microseconds summary
http_request_duration_microseconds{handler="webhook",quantile="0.5"} 2.2119355848e+07
http_request_duration_microseconds{handler="webhook",quantile="0.9"} 2.2119355848e+07
http_request_duration_microseconds{handler="webhook",quantile="0.99"} 2.2119355848e+07
http_request_duration_microseconds_sum{handler="webhook"} 2.2119355848e+07
http_request_duration_microseconds_count{handler="webhook"} 1
# HELP http_request_size_bytes The HTTP request sizes in bytes.
# TYPE http_request_size_bytes summary
http_request_size_bytes{handler="webhook",quantile="0.5"} 20248
http_request_size_bytes{handler="webhook",quantile="0.9"} 20248
http_request_size_bytes{handler="webhook",quantile="0.99"} 20248
http_request_size_bytes_sum{handler="webhook"} 20248
http_request_size_bytes_count{handler="webhook"} 1
# HELP http_requests_total Total number of HTTP requests made.
# TYPE http_requests_total counter
http_requests_total{code="200",handler="webhook",method="post"} 1
# HELP http_response_size_bytes The HTTP response sizes in bytes.
# TYPE http_response_size_bytes summary
http_response_size_bytes{handler="webhook",quantile="0.5"} 1293
http_response_size_bytes{handler="webhook",quantile="0.9"} 1293
http_response_size_bytes{handler="webhook",quantile="0.99"} 1293
http_response_size_bytes_sum{handler="webhook"} 1293
http_response_size_bytes_count{handler="webhook"} 1
```

## Future Work

- Increase logging to help debug issues
