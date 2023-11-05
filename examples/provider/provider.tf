# Configure the ksqlDB provider
terraform {
  required_providers {
    ksqldb = {
      source  = "shmyer/ksqldb"
      version = "0.3.0"
    }
  }
}

provider "ksqldb" {
  url      = var.ksqldb_url      # optionally use KSQLDB_URL environment variable
  username = var.ksqldb_username # optionally use KSQLDB_USERNAME environment variable
  password = var.ksqldb_password # optionally use KSQLDB_PASSWORD environment variable
}
# Create the resources
