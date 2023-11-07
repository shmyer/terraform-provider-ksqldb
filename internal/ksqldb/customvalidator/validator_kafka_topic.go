package customvalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"regexp"
)

var _ validator.String = kafkaTopicValidator{}

// kafkaTopicValidator validates that an identifier is valid.
type kafkaTopicValidator struct {
}

// Description describes the validation in plain text formatting.
func (v kafkaTopicValidator) Description(_ context.Context) string {
	return "topic name must be at max 255 characters and must only contain valid characters"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v kafkaTopicValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString performs the validation.
func (v kafkaTopicValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {

	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	if len(value) > 255 {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			v.Description(ctx),
			fmt.Sprintf("The topic name '%s' is too long. Must be up to 255 characters in length.", value),
		))
	}

	// others only allow letters, numbers and underscore
	matches, _ := regexp.MatchString("^[a-zA-Z0-9._-]+$", value)
	if !matches {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			v.Description(ctx),
			fmt.Sprintf("The topic name '%s' is invalid. It can include the following characters: a-z, A-Z, 0-9, . (dot), _ (underscore), and - (dash).", value),
		))
	}
}

// KafkaTopic returns an AttributeValidator which ensures that any configured
// attribute value:
//
//   - Is at max 255 characters long.
//   - Only contains the following characters: a-z, A-Z, 0-9, . (dot), _ (underscore), and - (dash).
func KafkaTopic() validator.String {
	return kafkaTopicValidator{}
}
