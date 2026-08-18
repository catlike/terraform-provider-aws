// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectConnectGatewayAttachmentsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping Direct Connect acceptance test in short mode")
	}
	connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")
	dataSourceName := "data.aws_dx_gateway_attachments.test"
	gatewayName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	gatewayASN := acctest.RandIntRange(t, 64512, 65534)
	vifASN := acctest.RandIntRange(t, 64512, 65534)
	vlan := acctest.RandIntRange(t, 2049, 4094)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DirectConnectEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckPrivateVirtualInterfaceDestroy(ctx, t),
			testAccCheckGatewayDestroy(ctx, t),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayAttachmentsDataSourceConfig(connectionID, gatewayName, gatewayASN, vifASN, vlan),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "attachments.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "attachments.0.attachment_type", "PrivateVirtualInterface"),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachments.0.direct_connect_gateway_id", "aws_dx_gateway.test", names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachments.0.virtual_interface_id", "aws_dx_private_virtual_interface.test", names.AttrID),
				),
			},
		},
	})
}

func testAccGatewayAttachmentsDataSourceConfig(connectionID, gatewayName string, gatewayASN, vifASN, vlan int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  amazon_side_asn = %[3]d
  name            = %[2]q
}

resource "aws_dx_private_virtual_interface" "test" {
  address_family = "ipv4"
  bgp_asn        = %[4]d
  connection_id  = %[1]q
  dx_gateway_id  = aws_dx_gateway.test.id
  name           = %[2]q
  vlan           = %[5]d
}

data "aws_dx_gateway_attachments" "test" {
  direct_connect_gateway_id = aws_dx_gateway.test.id

  depends_on = [aws_dx_private_virtual_interface.test]
}
`, connectionID, gatewayName, gatewayASN, vifASN, vlan)
}
