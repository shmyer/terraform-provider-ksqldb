package customvalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"strings"
)

var _ validator.String = queryValidator{}

// queryValidator validates that a query is valid.
type queryValidator struct {
}

// Description describes the validation in plain text formatting.
func (v queryValidator) Description(_ context.Context) string {
	return "The query must be valid"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v queryValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString performs the validation.
func (v queryValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {

	if request.ConfigValue.IsNull() {
		return
	}

	var source basetypes.BoolValue
	request.Config.GetAttribute(ctx, path.Root("source"), &source)

	if source.ValueBool() {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			v.Description(ctx),
			"The query attribute can't be used alongside the source attribute",
		))
	}

	value := request.ConfigValue.ValueString()

	// prevent injection
	if strings.Contains(value, ";") {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			v.Description(ctx),
			"The query must not contain a semicolon",
		))
	}

	// ensure it is a select statement
	if !strings.HasPrefix(strings.ToUpper(value), "SELECT ") {
		// TODO
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			v.Description(ctx),
			"The query must start with the SELECT keyword",
		))
	}
}

// Query returns an AttributeValidator which ensures that any configured
// attribute value:
//
//   - The query does not contain a semicolon
//   - The query starts with "SELECT "
func Query() validator.String {
	return queryValidator{}
}
