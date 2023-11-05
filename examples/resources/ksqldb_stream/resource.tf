resource "ksqldb_stream" "input" {
  name         = "INPUT"
  kafka_topic  = "input"
  key_format   = "AVRO"
  value_format = "AVRO"
}

resource "ksqldb_stream" "input_timestamp" {
  name         = "INPUT_TIMESTAMP"
  kafka_topic  = "input"
  key_format   = "AVRO"
  value_format = "AVRO"
  timestamp    = "TIMESTAMP"
}

resource "ksqldb_stream" "input_timestamp_format" {
  name             = "INPUT_TIMESTAMP_FORMAT"
  kafka_topic      = "input"
  key_format       = "AVRO"
  value_format     = "AVRO"
  timestamp        = "STRING"
  timestamp_format = "yyyy-MM-dd''T''HH:mm:ssX"
}

resource "ksqldb_stream" "input_schema_inference" {
  name            = "INPUT_SCHEMA_INFERENCE"
  kafka_topic     = "input"
  key_format      = "AVRO"
  value_format    = "AVRO"
  key_schema_id   = 6
  value_schema_id = 7
}

resource "ksqldb_stream" "input_properties" {
  name         = "INPUT_PROPERTIES"
  kafka_topic  = "input"
  key_format   = "AVRO"
  value_format = "AVRO"
  properties   = {
    "auto.offset.reset" = "earliest"
  }
}

resource "ksqldb_stream" "input_source" {
  name         = "INPUT_SOURCE"
  kafka_topic  = "input"
  key_format   = "AVRO"
  value_format = "AVRO"
  source       = true
}
