// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"cmp"
	"context"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkDataSource("aws_dx_gateway_attachments", name="Gateway Attachments")
// @Region(global=true)
func newGatewayAttachmentsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &gatewayAttachmentsDataSource{}, nil
}

type gatewayAttachmentsDataSource struct {
	framework.DataSourceWithModel[gatewayAttachmentsDataSourceModel]
}

func (d *gatewayAttachmentsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"attachments": framework.DataSourceComputedListOfObjectAttribute[gatewayAttachmentModel](ctx),
			"direct_connect_gateway_id": schema.StringAttribute{
				Optional: true,
			},
			"virtual_interface_id": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (d *gatewayAttachmentsDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("direct_connect_gateway_id"),
			path.MatchRoot("virtual_interface_id"),
		),
	}
}

func (d *gatewayAttachmentsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data gatewayAttachmentsDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	attachments, err := findGatewayAttachments(ctx, d.Meta().DirectConnectClient(ctx), gatewayAttachmentsReadInput(data))
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	flattened, diags := flattenGatewayAttachments(ctx, attachments)
	smerr.AddEnrich(ctx, &response.Diagnostics, diags)
	if response.Diagnostics.HasError() {
		return
	}
	data.Attachments = flattened

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func gatewayAttachmentsReadInput(data gatewayAttachmentsDataSourceModel) *directconnect.DescribeDirectConnectGatewayAttachmentsInput {
	input := &directconnect.DescribeDirectConnectGatewayAttachmentsInput{}
	if !data.DirectConnectGatewayID.IsNull() && !data.DirectConnectGatewayID.IsUnknown() {
		input.DirectConnectGatewayId = data.DirectConnectGatewayID.ValueStringPointer()
	}
	if !data.VirtualInterfaceID.IsNull() && !data.VirtualInterfaceID.IsUnknown() {
		input.VirtualInterfaceId = data.VirtualInterfaceID.ValueStringPointer()
	}

	return input
}

func sortGatewayAttachments(attachments []awstypes.DirectConnectGatewayAttachment) {
	slices.SortFunc(attachments, func(a, b awstypes.DirectConnectGatewayAttachment) int {
		if order := cmp.Compare(aws.ToString(a.DirectConnectGatewayId), aws.ToString(b.DirectConnectGatewayId)); order != 0 {
			return order
		}
		return cmp.Compare(aws.ToString(a.VirtualInterfaceId), aws.ToString(b.VirtualInterfaceId))
	})
}

// nosemgrep:ci.semgrep.framework.manual-flattener-functions
func flattenGatewayAttachments(ctx context.Context, attachments []awstypes.DirectConnectGatewayAttachment) (fwtypes.ListNestedObjectValueOf[gatewayAttachmentModel], diag.Diagnostics) {
	sortGatewayAttachments(attachments)

	var result fwtypes.ListNestedObjectValueOf[gatewayAttachmentModel]
	diags := fwflex.Flatten(ctx, attachments, &result)
	return result, diags
}

func findGatewayAttachments(ctx context.Context, conn *directconnect.Client, input *directconnect.DescribeDirectConnectGatewayAttachmentsInput) ([]awstypes.DirectConnectGatewayAttachment, error) {
	input.MaxResults = aws.Int32(100)

	var attachments []awstypes.DirectConnectGatewayAttachment
	if err := describeDirectConnectGatewayAttachmentsPages(ctx, conn, input, func(page *directconnect.DescribeDirectConnectGatewayAttachmentsOutput, _ bool) bool {
		if page != nil {
			attachments = append(attachments, page.DirectConnectGatewayAttachments...)
		}
		return true
	}); err != nil {
		return nil, err
	}

	if attachments == nil {
		return []awstypes.DirectConnectGatewayAttachment{}, nil
	}

	return attachments, nil
}

type gatewayAttachmentsDataSourceModel struct {
	Attachments            fwtypes.ListNestedObjectValueOf[gatewayAttachmentModel] `tfsdk:"attachments"`
	DirectConnectGatewayID types.String                                            `tfsdk:"direct_connect_gateway_id"`
	VirtualInterfaceID     types.String                                            `tfsdk:"virtual_interface_id"`
}

type gatewayAttachmentModel struct {
	AttachmentState              fwtypes.StringEnum[awstypes.DirectConnectGatewayAttachmentState] `tfsdk:"attachment_state"`
	AttachmentType               fwtypes.StringEnum[awstypes.DirectConnectGatewayAttachmentType]  `tfsdk:"attachment_type"`
	DirectConnectGatewayID       types.String                                                     `tfsdk:"direct_connect_gateway_id"`
	StateChangeError             types.String                                                     `tfsdk:"state_change_error"`
	VirtualInterfaceID           types.String                                                     `tfsdk:"virtual_interface_id"`
	VirtualInterfaceOwnerAccount types.String                                                     `tfsdk:"virtual_interface_owner_account_id"`
	VirtualInterfaceRegion       types.String                                                     `tfsdk:"virtual_interface_region"`
}
