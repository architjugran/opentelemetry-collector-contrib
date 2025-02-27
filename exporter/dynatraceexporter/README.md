# Dynatrace Exporter

| Status                   |                  |
| ------------------------ |------------------|
| Stability                | [beta]           |
| Supported pipeline types | metrics          |
| Distributions            | [contrib], [AWS] |

The [Dynatrace](https://www.dynatrace.com/integrations/opentelemetry/) metrics exporter exports metrics to the [Metrics API v2](https://www.dynatrace.com/support/help/dynatrace-api/environment-api/metric-v2/post-ingest-metrics/)
using the [metrics ingestion protocol](https://www.dynatrace.com/support/help/how-to-use-dynatrace/metrics/metric-ingestion/metric-ingestion-protocol/).
This enables Dynatrace to receive metrics collected by the OpenTelemetry Collector.  
More information on exporting metrics to Dynatrace can be found in the
[Dynatrace documentation for OpenTelemetry metrics](https://www.dynatrace.com/support/help/shortlink/opentelemetry-metrics).

For ingesting traces (spans) into Dynatrace, use the generic OTLP/HTTP exporter shipped with the Collector.  
More information on exporting traces to Dynatrace can be found in the
[Dynatrace documentation for OpenTelemetry traces](https://www.dynatrace.com/support/help/extend-dynatrace/opentelemetry/opentelemetry-traces/opentelemetry-ingest).

> The requests sent to Dynatrace are authenticated using an API token mechanism documented [here](https://www.dynatrace.com/support/help/dynatrace-api/basics/dynatrace-api-authentication/).  
> Please review the Collector's [security
> documentation](https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/security-best-practices.md),
> which contains recommendations on securing sensitive information such as the
> API key required by this exporter.

## Requirements

You will either need a Dynatrace OneAgent (version 1.201 or higher) installed on the same host as the Collector; or a Dynatrace environment with version 1.202 or higher.

- Collector contrib minimum version: 0.18.0


## Getting Started

The Dynatrace exporter is enabled by adding a `dynatrace` entry to the `exporters` section of your config file.
All configurations are optional, but if an `endpoint` other than the OneAgent metric ingestion endpoint is specified then an `api_token` is required.
To see all available options, see [Advanced Configuration](#advanced-configuration) below.

> When using this exporter, it is strongly RECOMMENDED to configure the OpenTelemetry SDKs to export metrics 
> with DELTA temporality. If you are exporting Sum or Histogram metrics with CUMULATIVE temporality, read
> about possible limitations of this exporter [below](#considerations-when-exporting-cumulative-data-points).

### Running alongside Dynatrace OneAgent (preferred)

If you run the Collector on a host or VM that is monitored by the Dynatrace OneAgent then you only need to enable the exporter. No further configurations needed. The Dynatrace exporter will send all metrics to the OneAgent which will use its secure and load balanced connection to send the metrics to your Dynatrace SaaS or Managed environment.
Depending on your environment, you might have to enable metrics ingestion on the OneAgent first as described in the [Dynatrace documentation](https://www.dynatrace.com/support/help/how-to-use-dynatrace/metrics/metric-ingestion/ingestion-methods/opentelemetry/).

> Note: The name and identifier of the host running the Collector will be added as a dimension to every metric. If this is undesirable, then the output plugin may be used in standalone mode using the directions below.

```yaml
exporters:
  dynatrace:
    ## No options are required. By default, metrics will be exported via the OneAgent on the local host.
```

### Running standalone

If you run the Collector on a host or VM without a OneAgent you will need to configure the Metrics v2 API endpoint of your Dynatrace environment to send the metrics to as well as an API token.

Find out how to create a token in the [Dynatrace documentation](https://www.dynatrace.com/support/help/dynatrace-api/basics/dynatrace-api-authentication/) or navigate to **Access tokens** in your Dynatrace environment and create a token with the 'Ingest metrics' (`metrics.ingest`) scope enabled. It is recommended to limit token scope to only this permission.

The endpoint for the Dynatrace Metrics API v2 is:

* on Dynatrace Managed: `https://{your-domain}/e/{your-environment-id}/api/v2/metrics/ingest`
* on Dynatrace SaaS: `https://{your-environment-id}.live.dynatrace.com/api/v2/metrics/ingest`

More details can be found in the [Dynatrace documentation for the Metrics v2 API](https://www.dynatrace.com/support/help/dynatrace-api/environment-api/metric-v2/post-ingest-metrics/).

```yaml
exporters:
  dynatrace:
    ## If no OneAgent is running on the host, endpoint and api_token need to be set

    ## Dynatrace Metrics Ingest v2 endpoint to receive metrics
    endpoint: "https://{your-environment-id}.live.dynatrace.com/api/v2/metrics/ingest"

    ## API token is required if an endpoint is specified and should be restricted to the 'Ingest metrics' scope
    ## hard-coded for illustration only, should be read from a secure source
    api_token: "your API token here" 
```

You can learn more about how to use the Dynatrace API [here](https://www.dynatrace.com/support/help/dynatrace-api/).

### Metric Batching

Dynatrace recommends the use of the batch processor with a maximum batch size of 1000 metrics and a timeout between 10 and 60 seconds. Batches with more than 1000 metrics may be throttled by Dynatrace.

Full example:

 ```yaml
receivers:
  # Collect own metrics and export them to Dynatrace
  prometheus:
    config:
      scrape_configs:
        - job_name: 'otel-collector'
          scrape_interval: 10s
          static_configs:
            - targets: [ '0.0.0.0:8888' ]

processors:
  batch:
    # Batch size must be less than or equal to 1000
    send_batch_max_size: 1000
    timeout: 30s

exporters:
  dynatrace:
    # optional - Dimensions specified here will be included as a dimension on every exported metric
    #            unless that metric already has a dimension with the same key.
    default_dimensions:
      example_dimension: example value

    # optional - prefix will be prepended to each metric name in prefix.name format
    prefix: my_prefix

    endpoint: https://abc12345.live.dynatrace.com/api/v2/metrics/ingest
    # Token must at least have the Ingest metrics (metrics.ingest) permission
    api_token: my_api_token

service:
  pipelines:
    metrics:
      receivers: [prometheus]
      processors: [batch]
      exporters: [dynatrace]

 ```

## Advanced Configuration

Several helper files are leveraged to provide additional capabilities automatically:

- [HTTP settings](https://github.com/open-telemetry/opentelemetry-collector/blob/main/config/confighttp/README.md)
- [TLS and mTLS settings](https://github.com/open-telemetry/opentelemetry-collector/blob/main/config/configtls/README.md)
- [Queuing, retry and timeout settings](https://github.com/open-telemetry/opentelemetry-collector/blob/main/exporter/exporterhelper/README.md) except timeout which is handled by the HTTP settings

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
exporters:
  dynatrace:
    endpoint: https://ab12345.live.dynatrace.com
    api_token: <api token must have metrics.write permission>
    default_dimensions:
      example_dimension: example value
    prefix: my_prefix
    headers:
      - header1: value1
    read_buffer_size: 4000
    write_buffer_size: 4000
    timeout: 30s
    tls:
      insecure_skip_verify: false # (default=false)
    retry_on_failure:
      enabled: true
      initial_interval: 5s
      max_interval: 30s
      max_elapsed_time: 120s
    sending_queue:
      enabled: true
      num_consumers: 10
      queue_size: 5000
    resource_to_telemetry_conversion:
      enabled: false
service:
  extensions:
  pipelines:
    metrics:
      receivers: [otlp]
      exporters: [dynatrace]
```

### default_dimensions (Optional)

`default_dimensions` are included as dimensions on all exported metrics unless that metric already has a dimension with the same key.
`default_dimensions` is specified as a map of string key-value pairs.

### prefix (Optional)

Prefix is a string which will be used as the first part of a dot-separated metric key.
For example, if a metric with name `request_count` is prefixed with `my_service`, the resulting
metric key is `my_service.request_count`.

### headers (Optional)

Additional headers to be included with every outgoing http request.

### read_buffer_size (Optional)

Defines the buffer size to allocate to the HTTP client for reading the response.

Default: `4096`

### write_buffer_size (Optional)

Defines the buffer size to allocate to the HTTP client for writing the payload

Default: `4096`

### timeout (Optional)

Timeout specifies a time limit for requests made by this
Client. The timeout includes connection time, any
redirects, and reading the response body. The timer remains
running after Get, Head, Post, or Do return and will
interrupt reading of the Response.Body.

https://golang.org/pkg/net/http/#Client

Default: `0`

### tls.insecure_skip_verify (Optional)

Additionally you can configure TLS to be enabled but skip verifying the server's certificate chain.
This cannot be combined with `insecure` since `insecure` won't use TLS at all.
More details can be found in the collector readme for
[TLS and mTLS settings](https://github.com/open-telemetry/opentelemetry-collector/blob/main/config/configtls/README.md).

Default: `false`

### retry_on_failure.enabled (Optional)

Default: `true`

### retry_on_failure.initial_interval (Optional)

Time to wait after the first failure before retrying; ignored if enabled is false.

Default: `5s`

### retry_on_failure.max_interval (Optional)

The upper bound on backoff; ignored if enabled is false

Default: `30s`

### retry_on_failure.max_elapsed_time (Optional)

The maximum amount of time spent trying to send a batch; ignored if enabled is false.

Default: `120s`

### sending_queue.enabled (Optional)

Default: `true`

### sending_queue.num_consumers (Optional)

Number of consumers that dequeue batches; ignored if enabled is false.

Default: `10`

### sending_queue.queue_size (Optional)

Maximum number of batches kept in memory before data; ignored if enabled is false;
User should calculate this as `num_seconds * requests_per_second` where:

- `num_seconds` is the number of seconds to buffer in case of a backend outage
- `requests_per_second` is the average number of requests per seconds.

Default: `5000`

### resource_to_telemetry_conversion (Optional)

When `resource_to_telemetry_conversion.enabled` is set to `true`, all resource
attributes will be included as metric dimensions in Dynatrace in addition to the
attributes present on the metric data point.

Default: `false`

> :warning: **Please note** that the Dynatrace API has a limit of `50` attributes
> per metric data point and any data point which exceeds this limit will be dropped.

If you think you might exceed this limit, you should use the
[transform processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/transformprocessor)
to apply a filter, so only a select subset of your resource attributes are converted.

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
processors:
  transform:
    metrics:
      queries:
        - keep_keys(resource.attributes, "key1", "key2", "key3")
exporters:
  dynatrace:
    endpoint: https://ab12345.live.dynatrace.com
    api_token: <api token must have metrics.write permission>
    resource_to_telemetry_conversion:
      enabled: true
service:
  pipelines:
    metrics:
      receivers: [otlp]
      processors: [transform]
      exporters: [dynatrace]
```

### tags (Deprecated, Optional)

**Deprecated: Please use [default_dimensions](#default_dimensions-optional) instead**

# Temporality

If possible when configuring your SDK, use DELTA temporality for Counter, Asynchronous Counter, and Histogram metrics.
Use CUMULATIVE temporality for UpDownCounter and Asynchronous UpDownCounter metrics.
When using OpenTelemetry SDKs to gather metrics data, setting the
`OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE` environment variable to `delta`
should correctly set temporality for all metrics.
You can check the [spec compliance matrix](https://github.com/open-telemetry/opentelemetry-specification/blob/main/spec-compliance-matrix.md#environment-variables)
if you are unsure if the SDK you are using supports this configuration.
You can read more about this and other configurations at
[OpenTelemetry Metrics Exporter - OTLP](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/metrics/sdk_exporters/otlp.md#additional-configuration).

## Considerations when exporting Cumulative Data Points

Histogram metrics with CUMULATIVE temporality are NOT SUPPORTED and will NOT be exported.

When possible, Sum metrics should use DELTA temporality.
When receiving Sum metrics with CUMULATIVE temporality, this exporter performs CUMULATIVE to DELTA conversion.
This conversion can lead to missing or inconsistent data, as described below:

### First Data Points are dropped

Due to the conversion, the exporter will drop the first received data point
after a counter is created or reset as there is no previous data point to compare it to.
This can be circumvented by configuring the OpenTelemetry SDK to export DELTA values.

## Multi-instance collector deployment

In a multiple-instance deployment of the OpenTelemetry Collector, the conversion 
can produce inconsistent data unless it can be guaranteed that metrics from the 
same source are processed by the same collector instance. This can be circumvented 
by configuring the OpenTelemetry SDK to export DELTA values.

[beta]:https://github.com/open-telemetry/opentelemetry-collector#beta
[contrib]:https://github.com/open-telemetry/opentelemetry-collector-releases/tree/main/distributions/otelcol-contrib
[AWS]:https://aws-otel.github.io/docs/partners/dynatrace
