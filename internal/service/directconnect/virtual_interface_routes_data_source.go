// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkDataSource("aws_dx_virtual_interface_routes", name="Virtual Interface Routes")
func newVirtualInterfaceRoutesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &virtualInterfaceRoutesDataSource{}, nil
}

type virtualInterfaceRoutesDataSource struct {
	framework.DataSourceWithModel[virtualInterfaceRoutesDataSourceModel]
}

func (d *virtualInterfaceRoutesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"address_family": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AddressFamily](),
				Optional:   true,
			},
			"as_path": schema.ListAttribute{
				CustomType:  fwtypes.ListOfInt64Type,
				ElementType: types.Int64Type,
				Optional:    true,
			},
			"cidrs": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtMost(10),
				},
			},
			"communities": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Optional:    true,
			},
			"route_direction": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RouteDirection](),
				Optional:   true,
			},
			"routes": framework.DataSourceComputedListOfObjectAttribute[virtualInterfaceRoutesDataSourceRouteModel](ctx),
			"virtual_interface_id": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *virtualInterfaceRoutesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().DirectConnectClient(ctx)

	var data virtualInterfaceRoutesDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	input := directconnect.ListVirtualInterfaceRoutesInput{
		VirtualInterfaceId: flex.StringFromFramework(ctx, data.VirtualInterfaceID),
	}
	if hasVirtualInterfaceRouteFilters(data) {
		filters := awstypes.RouteFilters{}
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, data, &filters))
		if resp.Diagnostics.HasError() {
			return
		}

		input.Filters = &filters
	}

	routes, err := listVirtualInterfaceRoutes(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.VirtualInterfaceID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, routes, &data.Routes))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func hasVirtualInterfaceRouteFilters(data virtualInterfaceRoutesDataSourceModel) bool {
	return !data.AddressFamily.IsNull() || !data.AsPath.IsNull() || !data.Cidrs.IsNull() || !data.Communities.IsNull() || !data.RouteDirection.IsNull()
}

func listVirtualInterfaceRoutes(ctx context.Context, conn *directconnect.Client, input *directconnect.ListVirtualInterfaceRoutesInput) ([]awstypes.Route, error) {
	routes := make([]awstypes.Route, 0)
	err := listVirtualInterfaceRoutesPages(ctx, conn, input, func(page *directconnect.ListVirtualInterfaceRoutesOutput, lastPage bool) bool {
		routes = append(routes, page.Routes...)

		return !lastPage
	})

	return routes, err
}

type virtualInterfaceRoutesDataSourceModel struct {
	framework.WithRegionModel
	AddressFamily      fwtypes.StringEnum[awstypes.AddressFamily]                                  `tfsdk:"address_family"`
	AsPath             fwtypes.ListOfInt64                                                         `tfsdk:"as_path"`
	Cidrs              fwtypes.ListOfString                                                        `tfsdk:"cidrs"`
	Communities        fwtypes.ListOfString                                                        `tfsdk:"communities"`
	RouteDirection     fwtypes.StringEnum[awstypes.RouteDirection]                                 `tfsdk:"route_direction"`
	Routes             fwtypes.ListNestedObjectValueOf[virtualInterfaceRoutesDataSourceRouteModel] `tfsdk:"routes"`
	VirtualInterfaceID types.String                                                                `tfsdk:"virtual_interface_id"`
}

type virtualInterfaceRoutesDataSourceRouteModel struct {
	AddressFamily      fwtypes.StringEnum[awstypes.AddressFamily]                                          `tfsdk:"address_family"`
	AsPath             fwtypes.ListNestedObjectValueOf[virtualInterfaceRoutesDataSourceAsPathSegmentModel] `tfsdk:"as_path"`
	AwsLogicalDeviceId types.String                                                                        `tfsdk:"aws_logical_device_id"`
	Cidr               types.String                                                                        `tfsdk:"cidr"`
	Communities        fwtypes.ListOfString                                                                `tfsdk:"communities"`
	RouteDirection     fwtypes.StringEnum[awstypes.RouteDirection]                                         `tfsdk:"route_direction"`
	RouteInstalledAt   timetypes.RFC3339                                                                   `tfsdk:"route_installed_at"`
}

type virtualInterfaceRoutesDataSourceAsPathSegmentModel struct {
	Path     fwtypes.ListOfInt64                     `tfsdk:"path"`
	PathType fwtypes.StringEnum[awstypes.AsPathType] `tfsdk:"path_type"`
}
