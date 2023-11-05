package modifiers

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var description = "If the stream is a source stream, every modification requires a replacement"

func isSourceStream(ctx context.Context, state tfsdk.State, requiresReplace *bool) {

	var source basetypes.BoolValue
	state.GetAttribute(ctx, path.Root("source"), &source)

	*requiresReplace = source.ValueBool()
}

func isSourceStreamMap(ctx context.Context, req planmodifier.MapRequest, resp *mapplanmodifier.RequiresReplaceIfFuncResponse) {
	isSourceStream(ctx, req.State, &resp.RequiresReplace)
}

var RequiresReplaceIfIsSourceStreamMap = mapplanmodifier.RequiresReplaceIf(isSourceStreamMap, description, description)

func isSourceStreamString(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	isSourceStream(ctx, req.State, &resp.RequiresReplace)
}

var RequiresReplaceIfIsSourceStreamString = stringplanmodifier.RequiresReplaceIf(isSourceStreamString, description, description)

func isSourceStreamInt64(ctx context.Context, req planmodifier.Int64Request, resp *int64planmodifier.RequiresReplaceIfFuncResponse) {
	isSourceStream(ctx, req.State, &resp.RequiresReplace)
}

var RequiresReplaceIfIsSourceStreamInt64 = int64planmodifier.RequiresReplaceIf(isSourceStreamInt64, description, description)
