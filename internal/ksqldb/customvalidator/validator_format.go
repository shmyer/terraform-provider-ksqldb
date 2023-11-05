package customvalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"slices"
)

var _ validator.String = formatValidator{}

// formatValidator validates that an identifier is valid.
type formatValidator struct {
}

// Description describes the validation in plain text formatting.
func (v formatValidator) Description(_ context.Context) string {
	return fmt.Sprintf("format must be one of %v", getAllowedFormats())
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v formatValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString performs the validation.
func (v formatValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {

	if request.ConfigValue.IsNull() {
		return
	}

	value := request.ConfigValue.ValueString()
	expected := getAllowedFormats()

	if !slices.Contains(expected, value) {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			v.Description(ctx),
			fmt.Sprintf("Unsupported format '%s'", value),
		))
	}
}

// Format returns an AttributeValidator which ensures that any configured
// attribute value:
//
//   - Is one of the allowed formats
func Format() validator.String {
	return formatValidator{}
}

func getAllowedFormats() []string {
	return []string{"NONE", "DELIMITED", "JSON", "JSON_SR", "AVRO", "KAFKA", "PROTOBUF", "PROTOBUF_NOSR"}
}
