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

func TestAccDirectConnectVirtualInterfacesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping Direct Connect acceptance test in short mode")
	}
	connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")
	rName := fmt.Sprintf("tf-testacc-vifs-%s", acctest.RandString(t, 9))
	bgpASN := acctest.RandIntRange(t, 64512, 65534)
	vlan := acctest.RandIntRange(t, 2049, 4094)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DirectConnectEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateVirtualInterfaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualInterfacesDataSourceConfig_basic(connectionID, rName, bgpASN, vlan),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("virtual_interface_included", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccVirtualInterfacesDataSourceConfig_basic(connectionID, name string, bgpASN, vlan int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  amazon_side_asn = 64512
  name            = %[2]q
}

resource "aws_dx_private_virtual_interface" "test" {
  address_family = "ipv4"
  bgp_asn        = %[3]d
  connection_id  = %[1]q
  dx_gateway_id  = aws_dx_gateway.test.id
  name           = %[2]q
  vlan           = %[4]d
}

data "aws_dx_virtual_interfaces" "test" {
  depends_on = [aws_dx_private_virtual_interface.test]
}

output "virtual_interface_included" {
  value = contains([for vif in data.aws_dx_virtual_interfaces.test.virtual_interfaces : vif.id], aws_dx_private_virtual_interface.test.id)
}
`, connectionID, name, bgpASN, vlan)
}
