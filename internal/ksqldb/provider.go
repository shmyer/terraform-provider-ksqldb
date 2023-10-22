package ksqldb

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &KsqldbProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &KsqldbProvider{
			version: version,
		}
	}
}

// KsqldbProvider is the provider implementation.
type KsqldbProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type KsqldbProviderModel struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	Url      types.String `tfsdk:"url"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// Metadata returns the provider type name.
func (p *KsqldbProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ksqldb"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *KsqldbProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Optional: true,
			},
			"username": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

// Configure prepares a HashiCups API client for data sources and resources.
func (p *KsqldbProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Check environment variables
	url := os.Getenv("KSQLDB_URL")
	username := os.Getenv("KSQLDB_USERNAME")
	password := os.Getenv("KSQLDB_PASSWORD")

	var data KsqldbProviderModel

	// Read configuration data into model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	// Check configuration data, which should take precedence over
	// environment variable data, if found.
	if data.Url.ValueString() != "" {
		url = data.Url.ValueString()
	}
	if data.Username.ValueString() != "" {
		username = data.Username.ValueString()
	}
	if data.Password.ValueString() != "" {
		password = data.Password.ValueString()
	}

	if url == "" {
		resp.Diagnostics.AddError(
			"Missing URL Configuration",
			"While configuring the provider, the ksqlDB URL was not found in "+
				"the KSQLDB_URL environment variable or provider "+
				"configuration block url attribute.",
		)
		// Not returning early allows the logic to collect all errors.
	}

	// Create data/clients and persist to resp.DataSourceData and
	// resp.ResourceData as appropriate.
	client := NewClient(&url, &username, &password)
	resp.ResourceData = client
}

// Resources defines the resources implemented in the provider.
func (p *KsqldbProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewStreamResource,
	}
}

// DataSources defines the data sources implemented in the provider.
func (p *KsqldbProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
