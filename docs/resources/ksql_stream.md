---
page_title: "stream Resource - terraform-provider-ksqldb"
subcategory: ""
description: |-
  The stream resource allows you to configure a ksqlDB stream.
---

# Resource `ksqldb_stream`

The stream resource allows you to configure a ksqlDB stream.

## Example Usage

```terraform
resource "ksqldb_stream" "clickstream" {
  name        = "CLICKS"
  kafka_topic = "clickstream"

  key_format      = "AVRO"
  value_format    = "AVRO"
}
```

## Argument Reference

- `name` - (Required) Name of the stream.
- `kafka_topic` - (Optional) The name of the Kafka topic that backs the stream. If this isn't set, the name of the stream in upper case is used as the topic name.
- `key_format` - (Optional) The serialization format of the message key in the topic.
- `value_format` - (Optional) The serialization format of the message value in the topic.
- `key_schema_id` - (Optional) The schema ID of the key schema in Schema Registry. The schema is used for schema inference and data serialization.
- `value_schema_id` - (Optional) The schema ID of the value schema in Schema Registry. The schema is used for schema inference and data serialization.
- `timestamp` - (Optional) Sets a column within the stream's schema to be used as the default source of ROWTIME for any downstream queries.