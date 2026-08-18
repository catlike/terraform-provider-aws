// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const testVirtualInterfaceOwnerAccountID = "123456789012" // nosemgrep:ci.literal-12Digit-string-test-constant

func TestVirtualInterfacesDataSourceSchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dataSource := &virtualInterfacesDataSource{}
	var response datasource.SchemaResponse
	dataSource.Schema(ctx, datasource.SchemaRequest{}, &response)

	if _, ok := response.Schema.Attributes[names.AttrID]; ok {
		t.Fatal("schema unexpectedly exposes synthetic id")
	}
	if _, ok := response.Schema.Attributes["virtual_interfaces"]; !ok {
		t.Fatal("schema is missing virtual_interfaces")
	}
	if _, ok := response.Schema.Attributes[names.AttrFilter]; ok {
		t.Fatal("schema unexpectedly exposes filter")
	}

	virtualInterfaces, ok := response.Schema.Attributes["virtual_interfaces"].(datasourceschema.ListAttribute)
	if !ok {
		t.Fatalf("virtual_interfaces type = %T, want ListAttribute", response.Schema.Attributes["virtual_interfaces"])
	}
	virtualInterfaceObject, ok := virtualInterfaces.ElementType.(types.ObjectType)
	if !ok {
		t.Fatalf("virtual_interfaces element type = %T, want ObjectType", virtualInterfaces.ElementType)
	}
	for _, name := range []string{names.AttrARN, "aws_device", names.AttrID, "rate_limit", "route_filter_prefixes", names.AttrTags} {
		if _, ok := virtualInterfaceObject.AttrTypes[name]; !ok {
			t.Errorf("virtual interface schema is missing %q", name)
		}
	}
	for _, name := range []string{"auth_key", "bgp_peers", "customer_router_config"} {
		if _, ok := virtualInterfaceObject.AttrTypes[name]; ok {
			t.Errorf("virtual interface schema unexpectedly exposes %q", name)
		}
	}
}

func TestFindVirtualInterfacesPages(t *testing.T) {
	t.Parallel()

	client, transport := newVirtualInterfacesTestClient(t,
		`{"nextToken":"next-page","virtualInterfaces":[{"virtualInterfaceId":"dxvif-2"}]}`,
		`{"virtualInterfaces":[{"virtualInterfaceId":"dxvif-1"}]}`,
	)

	virtualInterfaces, err := findVirtualInterfaces(context.Background(), client, &directconnect.DescribeVirtualInterfacesInput{}, func(*awstypes.VirtualInterface) bool {
		return true
	})

	if err != nil {
		t.Fatalf("finding virtual interfaces: %s", err)
	}
	if got, want := len(virtualInterfaces), 2; got != want {
		t.Fatalf("virtual interface count = %d, want %d", got, want)
	}
	if got, want := aws.ToString(virtualInterfaces[0].VirtualInterfaceId), "dxvif-2"; got != want {
		t.Errorf("first virtual interface ID = %q, want %q", got, want)
	}
	if got, want := aws.ToString(virtualInterfaces[1].VirtualInterfaceId), "dxvif-1"; got != want {
		t.Errorf("second virtual interface ID = %q, want %q", got, want)
	}
	if got, want := len(transport.requests), 2; got != want {
		t.Fatalf("request count = %d, want %d", got, want)
	}
	if got, want := string(transport.requests[0]), `"maxResults":100`; !bytes.Contains(transport.requests[0], []byte(want)) {
		t.Errorf("first request %s does not contain %s", got, want)
	}
	if got, want := string(transport.requests[1]), `"nextToken":"next-page"`; !bytes.Contains(transport.requests[1], []byte(want)) {
		t.Errorf("second request %s does not contain %s", got, want)
	}
}

func TestFindVirtualInterfacesNoMatch(t *testing.T) {
	t.Parallel()

	client, _ := newVirtualInterfacesTestClient(t, `{"virtualInterfaces":[{"virtualInterfaceId":"dxvif-1"}]}`)

	virtualInterfaces, err := findVirtualInterfaces(context.Background(), client, &directconnect.DescribeVirtualInterfacesInput{}, func(*awstypes.VirtualInterface) bool {
		return false
	})

	if err != nil {
		t.Fatalf("finding virtual interfaces: %s", err)
	}
	if got := len(virtualInterfaces); got != 0 {
		t.Errorf("virtual interface count = %d, want 0", got)
	}
}

func TestFindVirtualInterfaceByIDNotFound(t *testing.T) {
	t.Parallel()

	client, _ := newVirtualInterfacesTestClient(t, `{"virtualInterfaces":[]}`)

	_, err := findVirtualInterfaceByID(context.Background(), client, "dxvif-missing")

	if err == nil {
		t.Fatal("expected not found error")
	}
	if !retry.NotFound(err) {
		t.Errorf("error = %s, want not found", err)
	}
}

func TestFlattenVirtualInterfacesEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	virtualInterfaces, diags := flattenVirtualInterfaces(ctx, nil, "aws", nil)
	if diags.HasError() {
		t.Fatal(diags.Errors())
	}
	if virtualInterfaces.IsNull() {
		t.Fatal("virtual interfaces is null, want known empty list")
	}

	var models []virtualInterfaceModel
	if diags := virtualInterfaces.ElementsAs(ctx, &models, false); diags.HasError() {
		t.Fatal(diags.Errors())
	}
	if got := len(models); got != 0 {
		t.Errorf("virtual interface count = %d, want 0", got)
	}
}

func TestFlattenVirtualInterfaces(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	virtualInterfaces, diags := flattenVirtualInterfaces(ctx, []awstypes.VirtualInterface{
		{
			AddressFamily:          awstypes.AddressFamilyIPv4,
			AmazonAddress:          aws.String("192.0.2.1/30"),
			AmazonSideAsn:          aws.Int64(64512),
			Asn:                    64513,
			AsnLong:                aws.Int64(4200000000),
			AwsDeviceV2:            aws.String("device-v2"),
			AwsLogicalDeviceId:     aws.String("logical-device"),
			ConnectionId:           aws.String("dxcon-1"),
			CustomerAddress:        aws.String("192.0.2.2/30"),
			DirectConnectGatewayId: aws.String("dx-gateway-1"),
			JumboFrameCapable:      aws.Bool(true),
			Location:               aws.String("EqDC2"),
			Mtu:                    aws.Int32(9001),
			OwnerAccount:           aws.String(testVirtualInterfaceOwnerAccountID),
			RateLimit:              aws.String("500Mbps"),
			Region:                 aws.String("us-east-1"),
			RouteFilterPrefixes: []awstypes.RouteFilterPrefix{
				{Cidr: aws.String("192.168.0.0/16")},
				{Cidr: aws.String("10.0.0.0/8")},
			},
			SiteLinkEnabled:       aws.Bool(true),
			Tags:                  []awstypes.Tag{{Key: aws.String("Name"), Value: aws.String("vif-2")}, {Key: aws.String("aws:system"), Value: aws.String("ignored")}},
			VirtualGatewayId:      aws.String("vgw-1"),
			VirtualInterfaceId:    aws.String("dxvif-2"),
			VirtualInterfaceName:  aws.String("vif-2"),
			VirtualInterfaceState: awstypes.VirtualInterfaceStateAvailable,
			VirtualInterfaceType:  aws.String("private"),
			Vlan:                  101,
		},
		{
			OwnerAccount:         aws.String(testVirtualInterfaceOwnerAccountID),
			Region:               aws.String("us-east-1"),
			VirtualInterfaceId:   aws.String("dxvif-1"),
			VirtualInterfaceName: aws.String("vif-1"),
		},
	}, "aws", nil)

	if diags.HasError() {
		t.Fatal(diags.Errors())
	}
	if virtualInterfaces.IsNull() {
		t.Fatal("virtual interfaces is null, want known empty or populated list")
	}

	var models []virtualInterfaceModel
	if diags := virtualInterfaces.ElementsAs(ctx, &models, false); diags.HasError() {
		t.Fatal(diags.Errors())
	}
	if got, want := len(models), 2; got != want {
		t.Fatalf("virtual interface count = %d, want %d", got, want)
	}
	if got, want := models[0].VirtualInterfaceId.ValueString(), "dxvif-1"; got != want {
		t.Errorf("first virtual interface ID = %q, want %q", got, want)
	}
	if got, want := models[1].VirtualInterfaceId.ValueString(), "dxvif-2"; got != want {
		t.Errorf("second virtual interface ID = %q, want %q", got, want)
	}
	if got, want := models[1].Arn.ValueString(), "arn:aws:directconnect:us-east-1:123456789012:dxvif/dxvif-2"; got != want {
		t.Errorf("ARN = %q, want %q", got, want)
	}
	if got, want := models[1].AwsDeviceV2.ValueString(), "device-v2"; got != want {
		t.Errorf("AWS device = %q, want %q", got, want)
	}
	if got, want := models[1].RateLimit.ValueString(), "500Mbps"; got != want {
		t.Errorf("rate limit = %q, want %q", got, want)
	}
	var routeFilterPrefixes []string
	if diags := models[1].RouteFilterPrefixes.ElementsAs(ctx, &routeFilterPrefixes, false); diags.HasError() {
		t.Fatal(diags.Errors())
	}
	if got, want := fmt.Sprint(routeFilterPrefixes), "[10.0.0.0/8 192.168.0.0/16]"; got != want {
		t.Errorf("route filter prefixes = %s, want %s", got, want)
	}
	if got, want := models[1].Tags.Elements()["Name"].(types.String).ValueString(), "vif-2"; got != want {
		t.Errorf("Name tag = %q, want %q", got, want)
	}
	if _, ok := models[1].Tags.Elements()["aws:system"]; ok {
		t.Error("AWS tag was not ignored")
	}
	if models[0].Tags.IsNull() {
		t.Error("empty tags is null, want known empty map")
	}
	if models[0].RouteFilterPrefixes.IsNull() {
		t.Error("empty route filter prefixes is null, want known empty list")
	}
	routeFilterPrefixes = nil
	if diags := models[0].RouteFilterPrefixes.ElementsAs(ctx, &routeFilterPrefixes, false); diags.HasError() {
		t.Fatal(diags.Errors())
	}
	if got := len(routeFilterPrefixes); got != 0 {
		t.Errorf("empty route filter prefix count = %d, want 0", got)
	}
}

type virtualInterfacesPagingTestRoundTripper struct {
	mu        sync.Mutex
	responses [][]byte
	requests  [][]byte
}

func (r *virtualInterfacesPagingTestRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.responses) == 0 {
		return nil, io.EOF
	}

	requestBody, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}
	request.Body = io.NopCloser(bytes.NewReader(requestBody))
	r.requests = append(r.requests, requestBody)

	body := r.responses[0]
	r.responses = r.responses[1:]

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/x-amz-json-1.1"}},
		Body:       io.NopCloser(bytes.NewReader(body)),
		Request:    request,
	}, nil
}

func newVirtualInterfacesTestClient(t *testing.T, responses ...string) (*directconnect.Client, *virtualInterfacesPagingTestRoundTripper) {
	t.Helper()

	bodies := make([][]byte, len(responses))
	for i, response := range responses {
		bodies[i] = []byte(response)
	}

	transport := &virtualInterfacesPagingTestRoundTripper{responses: bodies}
	client := directconnect.New(directconnect.Options{
		Credentials: credentials.NewStaticCredentialsProvider("AKID", "SECRET", ""),
		HTTPClient:  &http.Client{Transport: transport},
		Region:      "us-east-1",
		Retryer:     aws.NopRetryer{},
	})

	return client, transport
}
