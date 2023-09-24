// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ksqldb

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceKsqldbStream() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKsqldbStreamCreate,
		ReadContext:   resourceKsqldbStreamRead,
		UpdateContext: resourceKsqldbStreamUpdate,
		DeleteContext: resourceKsqldbStreamDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateIdentifier,
			},

			"kafka_topic": &schema.Schema{
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateKafkaTopic,
			},
			"key_format": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateSerializationFormat,
			},
			"value_format": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateSerializationFormat,
			},
			"key_schema_id": &schema.Schema{
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          -1,
				ValidateDiagFunc: validateSchemaId,
			},
			"value_schema_id": &schema.Schema{
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          -1,
				ValidateDiagFunc: validateSchemaId,
			},
			"timestamp": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateIdentifier,
			},
			"properties": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Default:  map[string]string{},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceKsqldbStreamCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	stream, err := c.createStream(d, false, false)
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := buildId(c.url, stream.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	// set id
	d.SetId(*id)

	resourceKsqldbStreamRead(ctx, d, m)

	return diags
}

func resourceKsqldbStreamRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Get("name").(string)

	stream, err := c.describe(name)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("kafka_topic", stream.Topic); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("key_format", stream.KeyFormat); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("value_format", stream.ValueFormat); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	diags = append(diags, setSchemaIds(d, stream.Statement, true, KeySchemaIdPattern)...)
	diags = append(diags, setSchemaIds(d, stream.Statement, false, ValueSchemaIdPattern)...)

	diags = append(diags, setTimestamp(diags, d, stream)...)

	// TODO find a way to read properties

	return diags
}

func resourceKsqldbStreamUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// update stream
	_, err := c.updateStream(d, false)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	// read stream again in order to refresh state
	diagsRead := resourceKsqldbStreamRead(ctx, d, m)
	diags = append(diags, diagsRead...)

	return diags
}

func resourceKsqldbStreamDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Get("name").(string)

	err := c.dropStream(name)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
