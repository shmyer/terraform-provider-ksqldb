variable "ksqldb_url" {
  type = string
}

variable "ksqldb_username" {
  type      = string
  sensitive = true
}

variable "ksqldb_password" {
  type      = string
  sensitive = true
}
