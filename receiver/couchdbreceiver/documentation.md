[comment]: <> (Code generated by mdatagen. DO NOT EDIT.)

# couchdbreceiver

## Metrics

These are the metrics available for this scraper.

| Name | Description | Unit | Type | Attributes |
| ---- | ----------- | ---- | ---- | ---------- |
| **couchdb.average_request_time** | The average duration of a served request. | ms | Gauge(Double) | <ul> </ul> |
| **couchdb.database.open** | The number of open databases. | {databases} | Sum(Int) | <ul> </ul> |
| **couchdb.database.operations** | The number of database operations. | {operations} | Sum(Int) | <ul> <li>operation</li> </ul> |
| **couchdb.file_descriptor.open** | The number of open file descriptors. | {files} | Sum(Int) | <ul> </ul> |
| **couchdb.httpd.bulk_requests** | The number of bulk requests. | {requests} | Sum(Int) | <ul> </ul> |
| **couchdb.httpd.requests** | The number of HTTP requests by method. | {requests} | Sum(Int) | <ul> <li>http.method</li> </ul> |
| **couchdb.httpd.responses** | The number of each HTTP status code. | {responses} | Sum(Int) | <ul> <li>http.status_code</li> </ul> |
| **couchdb.httpd.views** | The number of views read. | {views} | Sum(Int) | <ul> <li>view</li> </ul> |

**Highlighted metrics** are emitted by default. Other metrics are optional and not emitted by default.
Any metric can be enabled or disabled with the following scraper configuration:

```yaml
metrics:
  <metric_name>:
    enabled: <true|false>
```

## Resource attributes

| Name | Description | Type |
| ---- | ----------- | ---- |
| couchdb.node.name | The name of the node. | Str |

## Metric attributes

| Name | Description | Values |
| ---- | ----------- | ------ |
| http.method | An HTTP request method. | COPY, DELETE, GET, HEAD, OPTIONS, POST, PUT |
| http.status_code | An HTTP status code. |  |
| operation | The operation type. | writes, reads |
| view | The view type. | temporary_view_reads, view_reads |
