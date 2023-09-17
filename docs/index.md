---
page_title: "Provider: ksqlDB"
subcategory: ""
description: |-
Terraform provider for interacting with ksqlDB API.
---

# ksqlDB Provider

The ksqlDB provider is used to interact with a ksqlDB cluster. It lets users...

1. manage ksqlDB streams
2. manage ksqlDB source streams

Use the navigation to the left to read about the available resources.

## Example Usage

Do not keep your authentication password in HCL for production environments, use Terraform environment variables.

```terraform
provider "ksqldb" {
  url      = "http://localhost:8088"
  username = "education"
  password = "test123"
}
```

## Schema

### Optional

- **username** (String, Optional) Username to authenticate to ksqlDB API
- **password** (String, Optional) Password to authenticate to ksqlDB API
- **url** (String, Optional) ksqlDB API base URL