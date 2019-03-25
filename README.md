# Jfrog Xray MS Teams Connector

## Motivation

This was created as teams does not accept the JSON payload from xray.

What this does is takes the Xray payload and convert it to a teams card format and forwards it onto teams.

## Usage

**Docker**

```bash
docker run -d [-e HTTPS_PROXY=""] squ1d123/xray-teams-connector:0.0.1
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

## Future Work

- Increase logging to help debug issues
- Support flags to be able to incude or exclude certain fields
- Parameterize Xray base URLs
