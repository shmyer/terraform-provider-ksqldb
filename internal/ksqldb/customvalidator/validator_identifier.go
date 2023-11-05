package customvalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"regexp"
	"strings"
	"terraform-provider-ksqldb/internal/ksqldb/util"
)

var _ validator.String = identifierValidator{}

// identifierValidator validates that an identifier is valid.
type identifierValidator struct {
}

// Description describes the validation in plain text formatting.
func (v identifierValidator) Description(_ context.Context) string {
	return "identifier must only contain valid characters"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v identifierValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString performs the validation.
func (v identifierValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {

	if request.ConfigValue.IsNull() || len(request.ConfigValue.ValueString()) == 0 {
		return
	}

	value := request.ConfigValue.ValueString()

	if strings.Contains(value, ";") {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			v.Description(ctx),
			fmt.Sprintf("The identifier '%s' must not contain a semicolon.", value),
		))
	}

	// back ticked identifiers can use any characters
	if util.IsBackTicked(value) {
		return
	}

	// others only allow capital letters, numbers and underscore
	matched, _ := regexp.MatchString(`^[A-Z0-9_]+$`, value)
	if !matched {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			v.Description(ctx),
			fmt.Sprintf("The identifier '%s' must only contain uppercase letters, numbers or underscore if it is not enclosed by backticks.", value),
		))
	}
}

// Identifier returns an AttributeValidator which ensures that any configured
// attribute value:
//
//   - Does not contain a semicolon.
//   - Only contains letters, numbers or underscore when the identifier is not enclosed by backticks.
func Identifier() validator.String {
	return identifierValidator{}
}
