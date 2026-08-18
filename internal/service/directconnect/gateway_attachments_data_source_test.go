// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awsdirectconnect "github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	frameworkdatasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestGatewayAttachmentsDataSourceSchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	d := &gatewayAttachmentsDataSource{}
	var response frameworkdatasource.SchemaResponse
	d.Schema(ctx, frameworkdatasource.SchemaRequest{}, &response)

	if response.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", response.Diagnostics)
	}

	if attribute, ok := response.Schema.Attributes["direct_connect_gateway_id"].(schema.StringAttribute); !ok || !attribute.Optional {
		t.Fatal("expected direct_connect_gateway_id to be an optional string attribute")
	}
	if attribute, ok := response.Schema.Attributes["virtual_interface_id"].(schema.StringAttribute); !ok || !attribute.Optional {
		t.Fatal("expected virtual_interface_id to be an optional string attribute")
	}
	if attribute, ok := response.Schema.Attributes["attachments"].(schema.ListAttribute); !ok || !attribute.Computed {
		t.Fatal("expected attachments to be a computed list attribute")
	}

	validateGatewayAttachmentsConfig(t, ctx, response.Schema, map[string]tftypes.Value{
		"attachments":               tftypes.NewValue(response.Schema.Attributes["attachments"].GetType().TerraformType(ctx), nil),
		"direct_connect_gateway_id": tftypes.NewValue(tftypes.String, nil),
		"virtual_interface_id":      tftypes.NewValue(tftypes.String, nil),
	}, true)
	validateGatewayAttachmentsConfig(t, ctx, response.Schema, map[string]tftypes.Value{
		"attachments":               tftypes.NewValue(response.Schema.Attributes["attachments"].GetType().TerraformType(ctx), nil),
		"direct_connect_gateway_id": tftypes.NewValue(tftypes.String, "dxgw-123"),
		"virtual_interface_id":      tftypes.NewValue(tftypes.String, nil),
	}, false)
	validateGatewayAttachmentsConfig(t, ctx, response.Schema, map[string]tftypes.Value{
		"attachments":               tftypes.NewValue(response.Schema.Attributes["attachments"].GetType().TerraformType(ctx), nil),
		"direct_connect_gateway_id": tftypes.NewValue(tftypes.String, nil),
		"virtual_interface_id":      tftypes.NewValue(tftypes.String, "dxvif-123"),
	}, false)
}

func TestGatewayAttachmentsReadInput(t *testing.T) {
	t.Parallel()

	input := gatewayAttachmentsReadInput(gatewayAttachmentsDataSourceModel{
		DirectConnectGatewayID: types.StringValue("dxgw-123"),
		VirtualInterfaceID:     types.StringValue("dxvif-123"),
	})

	if got, want := aws.ToString(input.DirectConnectGatewayId), "dxgw-123"; got != want {
		t.Errorf("DirectConnectGatewayId = %q, want %q", got, want)
	}
	if got, want := aws.ToString(input.VirtualInterfaceId), "dxvif-123"; got != want {
		t.Errorf("VirtualInterfaceId = %q, want %q", got, want)
	}
}

func TestFindGatewayAttachmentsPages(t *testing.T) {
	t.Parallel()

	var requests []*http.Request
	conn := testGatewayAttachmentsClient(t, func(request *http.Request) *http.Response {
		requests = append(requests, request)
		input := gatewayAttachmentsRequestInput(t, request)
		if got, want := input["maxResults"], float64(100); got != want {
			t.Errorf("maxResults = %v, want %v", got, want)
		}
		if got, want := input["directConnectGatewayId"], "dxgw-123"; got != want {
			t.Errorf("directConnectGatewayId = %v, want %v", got, want)
		}
		if got, want := input["virtualInterfaceId"], "dxvif-123"; got != want {
			t.Errorf("virtualInterfaceId = %v, want %v", got, want)
		}

		if input["nextToken"] == nil {
			return testGatewayAttachmentsResponse(`{"directConnectGatewayAttachments":[{"attachmentState":"attached","attachmentType":"PrivateVirtualInterface","directConnectGatewayId":"dxgw-2","virtualInterfaceId":"dxvif-2"}],"nextToken":"next-token"}`)
		}
		if got, want := input["nextToken"], "next-token"; got != want {
			t.Errorf("nextToken = %v, want %v", got, want)
		}
		return testGatewayAttachmentsResponse(`{"directConnectGatewayAttachments":[{"attachmentState":"attached","attachmentType":"PrivateVirtualInterface","directConnectGatewayId":"dxgw-1","virtualInterfaceId":"dxvif-1"}]}`)
	})

	attachments, err := findGatewayAttachments(context.Background(), conn, &awsdirectconnect.DescribeDirectConnectGatewayAttachmentsInput{
		DirectConnectGatewayId: aws.String("dxgw-123"),
		VirtualInterfaceId:     aws.String("dxvif-123"),
	})
	if err != nil {
		t.Fatalf("finding gateway attachments: %v", err)
	}
	if got, want := len(requests), 2; got != want {
		t.Fatalf("request count = %d, want %d", got, want)
	}
	if got, want := len(attachments), 2; got != want {
		t.Fatalf("attachment count = %d, want %d", got, want)
	}
}

func TestFindGatewayAttachmentsEmpty(t *testing.T) {
	t.Parallel()

	conn := testGatewayAttachmentsClient(t, func(*http.Request) *http.Response {
		return testGatewayAttachmentsResponse(`{"directConnectGatewayAttachments":[]}`)
	})

	attachments, err := findGatewayAttachments(context.Background(), conn, &awsdirectconnect.DescribeDirectConnectGatewayAttachmentsInput{
		DirectConnectGatewayId: aws.String("dxgw-123"),
	})
	if err != nil {
		t.Fatalf("finding gateway attachments: %v", err)
	}
	if attachments == nil {
		t.Fatal("expected an empty, non-nil attachment list")
	}
	if got := len(attachments); got != 0 {
		t.Errorf("attachment count = %d, want 0", got)
	}
}

func TestFlattenGatewayAttachments(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	attachments, diags := flattenGatewayAttachments(ctx, []awstypes.DirectConnectGatewayAttachment{
		{
			AttachmentState:              awstypes.DirectConnectGatewayAttachmentStateAttached,
			AttachmentType:               awstypes.DirectConnectGatewayAttachmentTypePrivateVirtualInterface,
			DirectConnectGatewayId:       aws.String("dxgw-2"),
			StateChangeError:             aws.String("example error"),
			VirtualInterfaceId:           aws.String("dxvif-2"),
			VirtualInterfaceOwnerAccount: aws.String("123456789012"), // nosemgrep:ci.literal-12Digit-string-test-constant
			VirtualInterfaceRegion:       aws.String("us-east-1"),
		},
		{DirectConnectGatewayId: aws.String("dxgw-1"), VirtualInterfaceId: aws.String("dxvif-1")},
	})
	if diags.HasError() {
		t.Fatal(diags.Errors())
	}
	if attachments.IsNull() {
		t.Fatal("attachments is null, want known list")
	}

	var models []gatewayAttachmentModel
	if diags := attachments.ElementsAs(ctx, &models, false); diags.HasError() {
		t.Fatal(diags.Errors())
	}
	if got, want := len(models), 2; got != want {
		t.Fatalf("attachment count = %d, want %d", got, want)
	}
	if got, want := models[0].VirtualInterfaceID.ValueString(), "dxvif-1"; got != want {
		t.Errorf("first virtual interface ID = %q, want %q", got, want)
	}
	if got, want := models[1].VirtualInterfaceOwnerAccount.ValueString(), "123456789012"; got != want { // nosemgrep:ci.literal-12Digit-string-test-constant
		t.Errorf("owner account = %q, want %q", got, want)
	}
	if got, want := awstypes.DirectConnectGatewayAttachmentType(models[1].AttachmentType.ValueString()), awstypes.DirectConnectGatewayAttachmentTypePrivateVirtualInterface; got != want {
		t.Errorf("attachment type = %q, want %q", got, want)
	}
}

func TestFlattenGatewayAttachmentsEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	attachments, diags := flattenGatewayAttachments(ctx, []awstypes.DirectConnectGatewayAttachment{})
	if diags.HasError() {
		t.Fatal(diags.Errors())
	}
	if attachments.IsNull() {
		t.Fatal("attachments is null, want known empty list")
	}
}

func TestSortGatewayAttachments(t *testing.T) {
	t.Parallel()

	attachments := []awstypes.DirectConnectGatewayAttachment{
		{DirectConnectGatewayId: aws.String("dxgw-2"), VirtualInterfaceId: aws.String("dxvif-1")},
		{DirectConnectGatewayId: aws.String("dxgw-1"), VirtualInterfaceId: aws.String("dxvif-2")},
		{DirectConnectGatewayId: aws.String("dxgw-1"), VirtualInterfaceId: aws.String("dxvif-1")},
	}

	sortGatewayAttachments(attachments)

	got := make([]string, 0, len(attachments))
	for _, attachment := range attachments {
		got = append(got, aws.ToString(attachment.DirectConnectGatewayId)+"/"+aws.ToString(attachment.VirtualInterfaceId))
	}
	if want := []string{"dxgw-1/dxvif-1", "dxgw-1/dxvif-2", "dxgw-2/dxvif-1"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("sorted attachments = %v, want %v", got, want)
	}
}

func validateGatewayAttachmentsConfig(t *testing.T, ctx context.Context, s schema.Schema, values map[string]tftypes.Value, wantError bool) {
	t.Helper()

	request := frameworkdatasource.ValidateConfigRequest{
		Config: tfsdk.Config{
			Raw:    tftypes.NewValue(s.Type().TerraformType(ctx), values),
			Schema: &s,
		},
	}
	var response frameworkdatasource.ValidateConfigResponse
	for _, validator := range (&gatewayAttachmentsDataSource{}).ConfigValidators(ctx) {
		validator.ValidateDataSource(ctx, request, &response)
	}
	if got := response.Diagnostics.HasError(); got != wantError {
		t.Fatalf("validator error = %t, want %t: %v", got, wantError, response.Diagnostics)
	}
}

type gatewayAttachmentsRoundTripper func(*http.Request) *http.Response

func (f gatewayAttachmentsRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request), nil
}

func testGatewayAttachmentsClient(t *testing.T, handler func(*http.Request) *http.Response) *awsdirectconnect.Client {
	t.Helper()

	return awsdirectconnect.NewFromConfig(aws.Config{
		Region:      "us-west-2",
		Credentials: credentials.NewStaticCredentialsProvider("access-key", "secret-key", ""),
		HTTPClient:  &http.Client{Transport: gatewayAttachmentsRoundTripper(handler)},
	}, func(options *awsdirectconnect.Options) {
		options.BaseEndpoint = aws.String("https://directconnect.example.test")
	})
}

func gatewayAttachmentsRequestInput(t *testing.T, request *http.Request) map[string]any {
	t.Helper()

	body, err := io.ReadAll(request.Body)
	if err != nil {
		t.Fatal(err)
	}
	request.Body = io.NopCloser(strings.NewReader(string(body)))

	var input map[string]any
	if err := json.Unmarshal(body, &input); err != nil {
		t.Fatalf("decoding request body: %s", err)
	}
	return input
}

func testGatewayAttachmentsResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/x-amz-json-1.1"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
