// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ksqldb

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceKsqldbSourceStream() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKsqldbSourceStreamCreate,
		ReadContext:   resourceKsqldbSourceStreamRead,
		DeleteContext: resourceKsqldbSourceStreamDelete,
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
				ForceNew:         true,
				ValidateDiagFunc: validateKafkaTopic,
			},
			"key_format": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSerializationFormat,
			},
			"value_format": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSerializationFormat,
			},
			"key_schema_id": &schema.Schema{
				Type:             schema.TypeInt,
				Optional:         true,
				ForceNew:         true,
				Default:          -1,
				ValidateDiagFunc: validateSchemaId,
			},
			"value_schema_id": &schema.Schema{
				Type:             schema.TypeInt,
				Optional:         true,
				ForceNew:         true,
				Default:          -1,
				ValidateDiagFunc: validateSchemaId,
			},
			"timestamp": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateIdentifier,
			},
			"properties": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Default:  map[string]string{},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceKsqldbSourceStreamCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	stream, err := c.createStream(d, false, true)
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := buildId(c.url, stream.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	// set id
	d.SetId(*id)

	resourceKsqldbSourceStreamRead(ctx, d, m)

	return diags
}

func resourceKsqldbSourceStreamRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

func resourceKsqldbSourceStreamDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
