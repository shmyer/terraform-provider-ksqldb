---
page_title: "ksqldb_materialized_stream Resource - terraform-provider-ksqldb"
subcategory: "Stream"
description: |-
  Configure a ksqlDB materialized stream
---

# Resource `ksqldb_materialized_stream`

The materialized stream resource allows you to configure a ksqlDB materialized stream, which will stream the results of the specified query into the sink topic.
The new stream is said to be a persistent query and is derived from the stream specified in the query.

## Example Usage

```terraform
resource "ksqldb_materialized_stream" "clickstream_mozilla" {
  name        = "CLICKS_MOZILLA"
  query       = "SELECT * FROM CLICKSTREAM WHERE AGENT = 'Mozilla'"
  
  kafka_topic = "clickstream-mozilla"

  key_format      = "AVRO"
  value_format    = "AVRO"
}
```

## Argument Reference

- `name` - (Required) Name of the stream.
- `query` - (Required) The KSQL SELECT statement which this stream is materialized from.
- `kafka_topic` - (Optional) The name of the Kafka topic that backs the stream. If this isn't set, the name of the stream in upper case is used as the topic name.
- `key_format` - (Optional) The serialization format of the message key in the topic.
- `value_format` - (Optional) The serialization format of the message value in the topic.
- `key_schema_id` - (Optional) The schema ID of the key schema in Schema Registry. The schema is used for schema inference and data serialization.
- `value_schema_id` - (Optional) The schema ID of the value schema in Schema Registry. The schema is used for schema inference and data serialization.
- `timestamp` - (Optional) Sets a column within the stream's schema to be used as the default source of ROWTIME for any downstream queries.
- `properties` - (Optional) Map of string properties to set as the "streamsProperties" parameter when issuing the KSQL statement via REST. 