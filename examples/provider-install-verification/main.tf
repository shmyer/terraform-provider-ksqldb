terraform {
  required_providers {
    ksqldb = {
      source = "hashicorp.com/shmyer/ksqldb"
    }
  }
}

provider "ksqldb" {
  url = "http://localhost:8088"
}

resource "ksqldb_stream" "test" {
  name         = "TEST2"
  kafka_topic  = "test"
  key_format   = "AVRO"
  value_format = "AVRO"
}
