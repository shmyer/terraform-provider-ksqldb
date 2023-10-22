// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ksqldb

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"regexp"
	"strconv"
	"strings"
	"terraform-provider-ksqldb/internal/ksqldb/customvalidator"
)

var KeySchemaIdPattern, _ = regexp.Compile(`(?i)KEY_SCHEMA_ID\s*=\s*(\d+)`)
var ValueSchemaIdPattern, _ = regexp.Compile(`(?i)VALUE_SCHEMA_ID\s*=\s*(\d+)`)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &StreamResource{}
var _ resource.ResourceWithImportState = &StreamResource{}

func NewStreamResource() resource.Resource {
	return &StreamResource{}
}

// StreamResource defines the resource implementation.
type StreamResource struct {
	client *Client
}

type StreamResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	KafkaTopic    types.String `tfsdk:"kafka_topic"`
	KeyFormat     types.String `tfsdk:"key_format"`
	ValueFormat   types.String `tfsdk:"value_format"`
	KeySchemaId   types.Int64  `tfsdk:"key_schema_id"`
	ValueSchemaId types.Int64  `tfsdk:"value_schema_id"`
	Timestamp     types.String `tfsdk:"timestamp"`
	Query         types.String `tfsdk:"query"`
}

func (r *StreamResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stream"
}

func (r *StreamResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Ksqldb stream resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Stream id",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the stream",
				Required:            true,
				Validators: []validator.String{
					customvalidator.Identifier(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"kafka_topic": schema.StringAttribute{
				MarkdownDescription: "The name of the Kafka topic that backs the stream. If this isn't set, the name of the stream in upper case is used as the topic name.",
				Required:            true,
				Validators: []validator.String{
					customvalidator.KafkaTopic(),
				},
			},
			"key_format": schema.StringAttribute{
				MarkdownDescription: "The serialization format of the message key in the topic.",
				Optional:            true,
				Validators: []validator.String{
					customvalidator.Format(),
				},
			},
			"value_format": schema.StringAttribute{
				MarkdownDescription: "The serialization format of the message value in the topic.",
				Optional:            true,
				Validators: []validator.String{
					customvalidator.Format(),
				},
			},
			"key_schema_id": schema.Int64Attribute{
				MarkdownDescription: "The schema ID of the key schema in Schema Registry. The schema is used for schema inference and data serialization.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"value_schema_id": schema.Int64Attribute{
				MarkdownDescription: "The schema ID of the value schema in Schema Registry. The schema is used for schema inference and data serialization.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"timestamp": schema.StringAttribute{
				MarkdownDescription: "Name of the stream",
				Optional:            true,
				Validators: []validator.String{
					customvalidator.Identifier(),
				},
			},
			"query": schema.StringAttribute{
				MarkdownDescription: "The KSQL SELECT statement which this stream is materialized from",
				Optional:            true,
				Validators: []validator.String{
					customvalidator.Query(),
				},
			},
			//"properties": schema.MapAttribute{
			//	MarkdownDescription: "Map of string properties to set as the \"streamsProperties\" parameter when issuing the KSQL statement via REST",
			//	Optional:            true,
			//	ElementType:         types.StringType,
			//},
		},
	}
}

func (r *StreamResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ksqldb.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *StreamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StreamResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	stream, err := r.client.createStream(ctx, data, false, false)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(err.Error(), err.Error()))
		return
	}

	// set id
	data.Id = types.StringValue(stream.Name)

	err = doReadInternal(ctx, &data, r.client)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(err.Error(), err.Error()))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StreamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StreamResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := doReadInternal(ctx, &data, r.client)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(err.Error(), err.Error()))
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Read data: %v", data))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func doReadInternal(ctx context.Context, data *StreamResourceModel, client *Client) error {

	// use id instead of name here for import functionality
	name := data.Id.ValueString()

	stream, err := client.describe(ctx, name)
	if err != nil {
		return err
	}

	tflog.Info(ctx, fmt.Sprintf("Read stream: %v", stream))

	data.Name = types.StringValue(stream.Name)
	data.KafkaTopic = types.StringValue(stream.Topic)
	data.KeyFormat = types.StringValue(stream.KeyFormat)
	data.ValueFormat = types.StringValue(stream.ValueFormat)

	err = setSchemaIds(data, stream.Statement, true, KeySchemaIdPattern)
	if err != nil {
		return err
	}
	err = setSchemaIds(data, stream.Statement, false, ValueSchemaIdPattern)
	if err != nil {
		return err
	}

	setTimestamp(data, stream)

	return nil
}

func (r *StreamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StreamResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// update stream
	_, err := r.client.updateStream(ctx, data, false)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(err.Error(), err.Error()))
		return
	}

	// read stream again in order to refresh state
	err = doReadInternal(ctx, &data, r.client)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(err.Error(), err.Error()))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StreamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StreamResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()

	err := r.client.dropStream(ctx, name)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(err.Error(), err.Error()))
		return
	}
}

func (r *StreamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if strings.ToUpper(req.ID) != req.ID {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Invalid ID", "The ID must be specified in uppercase"))
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func setTimestamp(data *StreamResourceModel, stream *Source) {

	timestamp := stream.Timestamp

	// if received timestamp is nil or empty, set nil in state
	if &timestamp == nil || len(timestamp) == 0 {
		data.Timestamp = types.StringNull()
		return
	}

	// get position of the timestamp field in the statement
	statement := stream.Statement
	index := strings.Index(statement, timestamp)

	// check whether the timestamp field was specified with backticks in the statement
	// if so, set the timestamp field with added backticks
	if statement[index-1] == '`' && statement[index+len(timestamp)] == '`' {
		newTimestamp := "`" + timestamp + "`"
		data.Timestamp = types.StringValue(newTimestamp)
		return
	}

	// otherwise just take the timestamp value as is
	data.Timestamp = types.StringValue(timestamp)
}

func setSchemaIds(data *StreamResourceModel, statement string, isKey bool, pattern *regexp.Regexp) error {

	// value schema id can't be found in response json but must be parsed from ksql statement
	schemaIdMatches := pattern.FindStringSubmatch(statement)
	if len(schemaIdMatches) > 1 {
		if schemaId := schemaIdMatches[1]; &schemaId != nil {

			// convert the found id to integer
			if schemaIdInt, err := strconv.Atoi(schemaId); err == nil {

				// set in state object
				if isKey {
					data.KeySchemaId = types.Int64Value(int64(schemaIdInt))
				} else {
					data.ValueSchemaId = types.Int64Value(int64(schemaIdInt))
				}
			} else {
				return err
			}
		}
	} else {

		if isKey {
			data.KeySchemaId = types.Int64Null()
		} else {
			data.ValueSchemaId = types.Int64Null()
		}
	}

	return nil
}
