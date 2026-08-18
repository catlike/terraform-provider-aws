// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package directconnect

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @FrameworkDataSource("aws_dx_virtual_interfaces", name="Virtual Interfaces")
func newVirtualInterfacesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &virtualInterfacesDataSource{}, nil
}

type virtualInterfacesDataSource struct {
	framework.DataSourceWithModel[virtualInterfacesDataSourceModel]
}

func (d *virtualInterfacesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"virtual_interfaces": framework.DataSourceComputedListOfObjectAttribute[virtualInterfaceModel](ctx),
		},
	}
}

func (d *virtualInterfacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data virtualInterfacesDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().DirectConnectClient(ctx)
	virtualInterfaces, err := findVirtualInterfaces(ctx, conn, &directconnect.DescribeVirtualInterfacesInput{}, func(*awstypes.VirtualInterface) bool {
		return true
	})
	if err != nil && !retry.NotFound(err) {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	flattened, diags := flattenVirtualInterfaces(ctx, virtualInterfaces, d.Meta().Partition(ctx), d.Meta().IgnoreTagsConfig(ctx))
	smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
	if resp.Diagnostics.HasError() {
		return
	}
	data.VirtualInterfaces = flattened

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type virtualInterfacesDataSourceModel struct {
	framework.WithRegionModel
	VirtualInterfaces fwtypes.ListNestedObjectValueOf[virtualInterfaceModel] `tfsdk:"virtual_interfaces"`
}

// virtualInterfaceModel intentionally excludes authentication and router configuration
// material (auth_key, BGP peers, and customer_router_config). This inventory data
// source exposes only non-secret VIF metadata.
type virtualInterfaceModel struct {
	AddressFamily          fwtypes.StringEnum[awstypes.AddressFamily]         `tfsdk:"address_family"`
	AmazonAddress          types.String                                       `tfsdk:"amazon_address"`
	AmazonSideAsn          types.Int64                                        `tfsdk:"amazon_side_asn"`
	Arn                    types.String                                       `tfsdk:"arn" autoflex:"-"`
	Asn                    types.Int32                                        `tfsdk:"bgp_asn"`
	AsnLong                types.Int64                                        `tfsdk:"bgp_asn_long"`
	AwsDeviceV2            types.String                                       `tfsdk:"aws_device"`
	AwsLogicalDeviceId     types.String                                       `tfsdk:"aws_logical_device_id"`
	ConnectionId           types.String                                       `tfsdk:"connection_id"`
	CustomerAddress        types.String                                       `tfsdk:"customer_address"`
	DirectConnectGatewayId types.String                                       `tfsdk:"direct_connect_gateway_id"`
	JumboFrameCapable      types.Bool                                         `tfsdk:"jumbo_frame_capable"`
	Location               types.String                                       `tfsdk:"location"`
	Mtu                    types.Int32                                        `tfsdk:"mtu"`
	OwnerAccount           types.String                                       `tfsdk:"owner_account_id"`
	RateLimit              types.String                                       `tfsdk:"rate_limit"`
	RouteFilterPrefixes    fwtypes.ListOfString                               `tfsdk:"route_filter_prefixes" autoflex:"-"`
	SiteLinkEnabled        types.Bool                                         `tfsdk:"site_link_enabled"`
	Tags                   tftags.Map                                         `tfsdk:"tags" autoflex:"-"`
	VirtualGatewayId       types.String                                       `tfsdk:"virtual_gateway_id"`
	VirtualInterfaceId     types.String                                       `tfsdk:"id"`
	VirtualInterfaceName   types.String                                       `tfsdk:"name"`
	VirtualInterfaceState  fwtypes.StringEnum[awstypes.VirtualInterfaceState] `tfsdk:"state"`
	VirtualInterfaceType   types.String                                       `tfsdk:"type"`
	Vlan                   types.Int32                                        `tfsdk:"vlan"`
}

// nosemgrep:ci.semgrep.framework.manual-flattener-functions
func flattenVirtualInterfaces(ctx context.Context, apiObjects []awstypes.VirtualInterface, partition string, ignoreTagsConfig *tftags.IgnoreConfig) (fwtypes.ListNestedObjectValueOf[virtualInterfaceModel], diag.Diagnostics) {
	slices.SortFunc(apiObjects, func(a, b awstypes.VirtualInterface) int {
		return cmp.Compare(aws.ToString(a.VirtualInterfaceId), aws.ToString(b.VirtualInterfaceId))
	})

	virtualInterfaces := make([]virtualInterfaceModel, 0, len(apiObjects))
	var diags diag.Diagnostics

	for _, apiObject := range apiObjects {
		var model virtualInterfaceModel
		diags.Append(fwflex.Flatten(ctx, apiObject, &model)...)
		if diags.HasError() {
			return fwtypes.ListNestedObjectValueOf[virtualInterfaceModel]{}, diags
		}

		model.Arn = fwflex.StringValueToFramework(ctx, arn.ARN{
			Partition: partition,
			Region:    aws.ToString(apiObject.Region),
			Service:   "directconnect",
			AccountID: aws.ToString(apiObject.OwnerAccount),
			Resource:  fmt.Sprintf("dxvif/%s", aws.ToString(apiObject.VirtualInterfaceId)),
		}.String())
		model.Tags = tftags.NewMapFromMapValue(fwflex.FlattenFrameworkStringValueMapLegacy(ctx, keyValueTags(ctx, apiObject.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()))

		routeFilterPrefixes := make([]string, 0, len(apiObject.RouteFilterPrefixes))
		for _, routeFilterPrefix := range apiObject.RouteFilterPrefixes {
			routeFilterPrefixes = append(routeFilterPrefixes, aws.ToString(routeFilterPrefix.Cidr))
		}
		slices.Sort(routeFilterPrefixes)
		model.RouteFilterPrefixes = fwflex.FlattenFrameworkStringValueListOfStringLegacy(ctx, routeFilterPrefixes)

		virtualInterfaces = append(virtualInterfaces, model)
	}

	data, d := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, virtualInterfaces)
	diags.Append(d...)
	return data, diags
}
