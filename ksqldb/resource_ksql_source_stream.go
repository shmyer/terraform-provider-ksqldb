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
		ReadContext:   resourceKsqldbStreamRead,
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
		},
	}
}

func resourceKsqldbSourceStreamCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	stream, err := c.CreateStream(d, true)
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
