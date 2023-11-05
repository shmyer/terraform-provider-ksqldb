package ksqldb

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strings"
)

func createStreamKsql(ctx context.Context, name string, source bool, materialized bool, data StreamResourceModel) (*string, error) {

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

	appendIfSpecified(&sb, "KAFKA_TOPIC", data.KafkaTopic, &length, lengthBeforeProperties)
	appendIfSpecified(&sb, "PARTITIONS", data.Partitions, &length, lengthBeforeProperties)
	appendIfSpecified(&sb, "REPLICAS", data.Replicas, &length, lengthBeforeProperties)
	appendIfSpecified(&sb, "RETENTION_MS", data.Retention, &length, lengthBeforeProperties)
	appendIfSpecified(&sb, "TIMESTAMP", data.Timestamp, &length, lengthBeforeProperties)
	appendIfSpecified(&sb, "TIMESTAMP_FORMAT", data.TimestampFormat, &length, lengthBeforeProperties)
	appendIfSpecified(&sb, "KEY_FORMAT", data.KeyFormat, &length, lengthBeforeProperties)
	appendIfSpecified(&sb, "VALUE_FORMAT", data.ValueFormat, &length, lengthBeforeProperties)
	appendIfSpecified(&sb, "KEY_SCHEMA_ID", data.KeySchemaId, &length, lengthBeforeProperties)
	appendIfSpecified(&sb, "VALUE_SCHEMA_ID", data.ValueSchemaId, &length, lengthBeforeProperties)

	sb.WriteString(")")

	if materialized {
		query := data.Query.ValueString()

		sb.WriteString(" AS ")
		sb.WriteString(query)
	}

	sb.WriteString(";")

	ksql := sb.String()

	tflog.Info(ctx, fmt.Sprintf("Created KSQL statement: %s", ksql))

	return &ksql, nil
}

func appendIfSpecified(sb *strings.Builder, property string, value attr.Value, length *int, lengthBeforeProperties int) {

	if value == nil || value.IsNull() || value.IsUnknown() {
		return
	}

	if *length > lengthBeforeProperties {
		sb.WriteString(", ")
	}

	if s, ok := value.(types.String); ok {
		sb.WriteString(fmt.Sprintf("%s = '%s'", property, s.ValueString()))
	} else if i, ok := value.(types.Int64); ok {
		sb.WriteString(fmt.Sprintf("%s = '%d'", property, i.ValueInt64()))
	}

	*length = sb.Len()
}
