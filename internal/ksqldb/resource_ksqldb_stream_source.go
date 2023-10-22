// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ksqldb

//
//import (
//	"context"
//	"fmt"
//	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
//	"github.com/hashicorp/terraform-plugin-framework/resource"
//	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
//	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
//	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
//	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
//	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
//	"github.com/hashicorp/terraform-plugin-framework/types"
//	"terraform-provider-ksqldb/internal/ksqldb/customvalidator"
//)
//
//// Ensure provider defined types fully satisfy framework interfaces.
//var _ resource.Resource = &SourceStreamResource{}
//var _ resource.ResourceWithImportState = &SourceStreamResource{}
//
//func NewSourceStreamResource() resource.Resource {
//	return &SourceStreamResource{}
//}
//
//// SourceStreamResource defines the resource implementation.
//type SourceStreamResource struct {
//	client *Client
//}
//
//// SourceStreamResourceModel describes the resource data model.
//type SourceStreamResourceModel struct {
//	Id            types.String `tfsdk:"id"`
//	Name          types.String `tfsdk:"name"`
//	KafkaTopic    types.String `tfsdk:"kafka_topic"`
//	KeyFormat     types.String `tfsdk:"key_format"`
//	ValueFormat   types.String `tfsdk:"value_format"`
//	KeySchemaId   types.Int64  `tfsdk:"key_schema_id"`
//	ValueSchemaId types.Int64  `tfsdk:"value_schema_id"`
//	Timestamp     types.String `tfsdk:"timestamp"`
//}
//
//func (r *SourceStreamResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
//	resp.TypeName = req.ProviderTypeName + "_source_stream"
//}
//
//func (r *SourceStreamResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
//	resp.Schema = schema.Schema{
//		// This description is used by the documentation generator and the language server.
//		MarkdownDescription: "Ksqldb source stream resource",
//
//		Attributes: map[string]schema.Attribute{
//			"id": schema.StringAttribute{
//				Computed:            true,
//				MarkdownDescription: "Stream id",
//				PlanModifiers: []planmodifier.String{
//					stringplanmodifier.UseStateForUnknown(),
//				},
//			},
//
//			"name": schema.StringAttribute{
//				MarkdownDescription: "Name of the stream",
//				Required:            true,
//				Validators: []validator.String{
//					customvalidator.Identifier(),
//				},
//				PlanModifiers: []planmodifier.String{
//					stringplanmodifier.RequiresReplace(),
//				},
//			},
//
//			"kafka_topic": schema.StringAttribute{
//				MarkdownDescription: "The name of the Kafka topic that backs the stream. If this isn't set, the name of the stream in upper case is used as the topic name.",
//				Required:            true,
//				Validators: []validator.String{
//					customvalidator.KafkaTopic(),
//				},
//				PlanModifiers: []planmodifier.String{
//					stringplanmodifier.RequiresReplace(),
//				},
//			},
//			"key_format": schema.StringAttribute{
//				MarkdownDescription: "The serialization format of the message key in the topic.",
//				Optional:            true,
//				Validators: []validator.String{
//					customvalidator.Format(),
//				},
//				PlanModifiers: []planmodifier.String{
//					stringplanmodifier.RequiresReplace(),
//				},
//			},
//			"value_format": schema.StringAttribute{
//				MarkdownDescription: "The serialization format of the message value in the topic.",
//				Optional:            true,
//				Validators: []validator.String{
//					customvalidator.Format(),
//				},
//				PlanModifiers: []planmodifier.String{
//					stringplanmodifier.RequiresReplace(),
//				},
//			},
//			"key_schema_id": schema.Int64Attribute{
//				MarkdownDescription: "The schema ID of the key schema in Schema Registry. The schema is used for schema inference and data serialization.",
//				Optional:            true,
//				Validators: []validator.Int64{
//					int64validator.AtLeast(1),
//				},
//				PlanModifiers: []planmodifier.Int64{
//					int64planmodifier.RequiresReplace(),
//				},
//			},
//			"value_schema_id": schema.Int64Attribute{
//				MarkdownDescription: "The schema ID of the value schema in Schema Registry. The schema is used for schema inference and data serialization.",
//				Optional:            true,
//				Validators: []validator.Int64{
//					int64validator.AtLeast(1),
//				},
//				PlanModifiers: []planmodifier.Int64{
//					int64planmodifier.RequiresReplace(),
//				},
//			},
//			"timestamp": schema.StringAttribute{
//				MarkdownDescription: "Name of the stream",
//				Optional:            true,
//				Validators: []validator.String{
//					customvalidator.Identifier(),
//				},
//				PlanModifiers: []planmodifier.String{
//					stringplanmodifier.RequiresReplace(),
//				},
//			},
//			//"properties": schema.MapAttribute{
//			//	MarkdownDescription: "Map of string properties to set as the \"streamsProperties\" parameter when issuing the KSQL statement via REST",
//			//	Optional:            true,
//			//	ElementType:         types.StringType,
//			//	PlanModifiers: []planmodifier.Map{
//			//		mapplanmodifier.RequiresReplace(),
//			//	},
//			//},
//		},
//	}
//}
//
//func (r *SourceStreamResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
//	// Prevent panic if the provider has not been configured.
//	if req.ProviderData == nil {
//		return
//	}
//
//	client, ok := req.ProviderData.(*Client)
//
//	if !ok {
//		resp.Diagnostics.AddError(
//			"Unexpected Resource Configure Type",
//			fmt.Sprintf("Expected *ksqldb.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
//		)
//
//		return
//	}
//
//	r.client = client
//}
//
//func (r *SourceStreamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
//	doCreate(ctx, r.client, req, resp, false, true)
//}
//
//func (r *SourceStreamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
//	doRead(ctx, r.client, req, resp)
//}
//
//func (r *SourceStreamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
//	// not supported
//}
//
//func (r *SourceStreamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
//	doDelete(ctx, r.client, req, resp)
//}
//
//func (r *SourceStreamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
//	doImportState(ctx, r.client, req, resp)
//}
