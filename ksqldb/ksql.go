package ksqldb

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
)

func createStreamKsql(name string, source bool, materialized bool, d *schema.ResourceData) (*string, error) {

	kafkaTopic := d.Get("kafka_topic").(string)
	keyFormat := d.Get("key_format").(string)
	valueFormat := d.Get("value_format").(string)
	keySchemaId := d.Get("key_schema_id").(int)
	valueSchemaId := d.Get("value_schema_id").(int)

	var sb strings.Builder

	sb.WriteString("CREATE")

	if source {
		sb.WriteString(" SOURCE")
	} else {
		sb.WriteString(" OR REPLACE")
	}

	sb.WriteString(" STREAM ")
	sb.WriteString(name)
	sb.WriteString(" WITH (")

	lengthBeforeProperties := sb.Len()
	length := sb.Len()

	appendIfSpecified(&sb, "KAFKA_TOPIC", kafkaTopic, &length, lengthBeforeProperties)
	appendIfSpecified(&sb, "KEY_FORMAT", keyFormat, &length, lengthBeforeProperties)
	appendIfSpecified(&sb, "VALUE_FORMAT", valueFormat, &length, lengthBeforeProperties)
	appendIfSpecified(&sb, "KEY_SCHEMA_ID", keySchemaId, &length, lengthBeforeProperties)
	appendIfSpecified(&sb, "VALUE_SCHEMA_ID", valueSchemaId, &length, lengthBeforeProperties)

	sb.WriteString(")")

	if materialized {
		query := d.Get("query").(string)

		sb.WriteString(" AS ")
		sb.WriteString(query)
	}

	sb.WriteString(";")

	ksql := sb.String()

	return &ksql, nil
}

func appendIfSpecified(sb *strings.Builder, property string, value interface{}, length *int, lengthBeforeProperties int) {

	if value == nil {
		return
	}

	if i, ok := value.(int); ok && i == -1 {
		return
	}

	if *length > lengthBeforeProperties {
		sb.WriteString(", ")
	}

	sb.WriteString(fmt.Sprintf("%s = '%v'", property, value))

	*length = sb.Len()
}
