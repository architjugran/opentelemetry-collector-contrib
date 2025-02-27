# Authenticator - Bearer

| Status                   |                      |
|--------------------------|----------------------|
| Stability                | [beta]               |
| Distributions            | [contrib]            |


This extension implements `configauth.ClientAuthenticator` and can be used in both http and gRPC receivers inside the `auth` settings, as a means to embed a static token for every RPC call that will be made.

The authenticator type has to be set to `bearertokenauth`.

## Configuration

- `scheme`: Specifies the auth scheme name. Defaults to "Bearer". Optional.

- `token`: Static authorization token that needs to be sent on every gRPC client call as metadata.

- `filename`: Name of file that contains a authorization token that needs to be sent in every client call.

Either one of `token` or `filename` field is required. If both are specified, then the `token` field value is **ignored**. In any case, the value of the token will be prepended by `${scheme}` before being sent as a value of "authorization" key in the request header in case of HTTP and metadata in case of gRPC.

**Note**: bearertokenauth requires transport layer security enabled on the exporter.


```yaml
extensions:
  bearertokenauth:
    token: "somerandomtoken"
    filename: "file-containing.token"
  bearertokenauth/withscheme:
    scheme: "Bearer"
    token: "randomtoken"

receivers:
  hostmetrics:
    scrapers:
      memory:
  otlp:
    protocols:
      grpc:

exporters:
  otlp/withauth:
    endpoint: 0.0.0.0:5000
    ca_file: /tmp/certs/ca.pem
    auth:
      authenticator: bearertokenauth

  otlphttp/withauth:
    endpoint: http://localhost:9000
    auth:
      authenticator: bearertokenauth/withscheme

service:
  extensions: [bearertokenauth, bearertokenauth/withscheme]
  pipelines:
    metrics:
      receivers: [hostmetrics]
      processors: []
      exporters: [otlp/withauth, otlphttp/withauth]
```


[beta]:https://github.com/open-telemetry/opentelemetry-collector#beta
[contrib]:https://github.com/open-telemetry/opentelemetry-collector-releases/tree/main/distributions/otelcol-contrib
