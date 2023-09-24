package ksqldb

import (
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neturl "net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var KeySchemaIdPattern, _ = regexp.Compile(`(?i)KEY_SCHEMA_ID\s*=\s*(\d+)`)
var ValueSchemaIdPattern, _ = regexp.Compile(`(?i)VALUE_SCHEMA_ID\s*=\s*(\d+)`)

func getAllowedSerializationFormats() []string {
	return []string{"NONE", "DELIMITED", "JSON", "JSON_SR", "AVRO", "KAFKA", "PROTOBUF", "PROTOBUF_NOSR"}
}

func setTimestamp(diags diag.Diagnostics, d *schema.ResourceData, stream *Source) diag.Diagnostics {

	timestamp := stream.Timestamp

	// if received timestamp is nil or empty, set nil in state
	if &timestamp == nil || len(timestamp) == 0 {
		if err := d.Set("timestamp", nil); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
		return diags
	}

	// get position of the timestamp field in the statement
	statement := stream.Statement
	index := strings.Index(statement, timestamp)

	// check whether the timestamp field was specified with backticks in the statement
	// if so, set the timestamp field with added backticks
	if statement[index-1] == '`' && statement[index+len(timestamp)] == '`' {
		newTimestamp := "`" + timestamp + "`"
		if err := d.Set("timestamp", newTimestamp); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
		return diags
	}

	// otherwise just take the timestamp value as is
	if err := d.Set("timestamp", timestamp); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	return diags
}

func setSchemaIds(d *schema.ResourceData, statement string, isKey bool, pattern *regexp.Regexp) diag.Diagnostics {

	attribute := "value_schema_id"
	if isKey {
		attribute = "key_schema_id"
	}

	var diags diag.Diagnostics

	// value schema id can't be found in response json but must be parsed from ksql statement
	schemaIdMatches := pattern.FindStringSubmatch(statement)
	if len(schemaIdMatches) > 1 {
		if schemaId := schemaIdMatches[1]; &schemaId != nil {

			// convert the found id to integer
			if schemaIdInt, err := strconv.Atoi(schemaId); err == nil {

				// set in state object
				if err := d.Set(attribute, schemaIdInt); err != nil {
					diags = append(diags, diag.FromErr(err)...)
				}
			} else {
				diags = append(diags, diag.FromErr(err)...)
			}
		}
	} else {

		if err := d.Set(attribute, -1); err != nil {

			// when there was no match the id isn't specified and must be set to -1
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	return diags
}

func buildId(url string, name string) (*string, error) {

	parsed, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("%s/%s", parsed.Hostname(), name)

	return &id, nil
}

func validateQuery(v any, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	value := v.(string)

	// prevent injection
	if strings.Contains(value, ";") {
		d := diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "The query contains invalid characters",
			Detail:   "The query must not contain a semicolon.",
		}
		diags = append(diags, d)
	}

	// ensure it is a select statement
	if !strings.HasPrefix(strings.ToUpper(value), "SELECT ") {
		d := diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "The query must start with the SELECT keyword",
		}
		diags = append(diags, d)
	}

	return diags
}

func validateIdentifier(v any, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	value := v.(string)

	if strings.Contains(value, ";") {
		d := diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "The identifier contains invalid characters",
			Detail:   fmt.Sprintf("The identifier '%s' must not contain a semicolon.", value),
		}
		return append(diags, d)
	}

	// back ticked identifiers can use any characters
	if isBackTicked(value) {
		return diags
	}

	// others only allow letters, numbers and underscore
	matched, _ := regexp.MatchString(`^(?i)[a-z0-9_]+$`, value)
	if !matched {
		d := diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "The identifier contains invalid characters",
			Detail:   fmt.Sprintf("The identifier '%s' must only contain letters, numbers or underscore.", value),
		}
		diags = append(diags, d)
	}

	return diags
}

func isBackTicked(v string) bool {
	return v[0] == '`' && v[len(v)-1] == '`'
}

func validateKafkaTopic(v any, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	value := v.(string)

	if len(value) > 255 {
		d := diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "The topic name is too long",
			Detail:   fmt.Sprintf("The topic name can be up to 255 characters in length."),
		}
		diags = append(diags, d)
	}

	matches, _ := regexp.MatchString("^[a-zA-Z0-9._-]+$", value)
	if !matches {
		d := diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "The topic name contains invalid characters",
			Detail:   fmt.Sprintf("The topic name can include the following characters: a-z, A-Z, 0-9, . (dot), _ (underscore), and - (dash)."),
		}
		diags = append(diags, d)
	}

	return diags
}

func validateSerializationFormat(v any, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	value := v.(string)
	expected := getAllowedSerializationFormats()

	if !slices.Contains(expected, value) {
		d := diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unsupported serialization format '%s'", value),
			Detail:   fmt.Sprintf("The serialization format must be one of %v", expected),
		}
		diags = append(diags, d)
	}

	return diags
}

func validateSchemaId(v any, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	value := v.(int)

	if value < 0 {
		d := diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "The schema id must be positive",
		}
		diags = append(diags, d)
	}

	return diags
}
