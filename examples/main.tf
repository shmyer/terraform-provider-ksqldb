# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_providers {
    ksqldb = {
      version = "0.2.0"
      source  = "shmyer/ksqldb"
    }
  }
}

provider "ksqldb" {}

resource "ksqldb_source_stream" "clickstream" {
  name        = "CLICKS"
  kafka_topic = "clickstream"

  key_format      = "AVRO"
  value_format    = "AVRO"
}