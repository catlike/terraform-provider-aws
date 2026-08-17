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

func TestAccDirectConnectVirtualInterfaceRoutesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	virtualInterfaceID := acctest.SkipIfEnvVarNotSet(t, "DX_VIRTUAL_INTERFACE_ID")
	dataSourceName := "data.aws_dx_virtual_interface_routes.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DirectConnectEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualInterfaceRoutesDataSourceConfig_basic(virtualInterfaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "routes.0.address_family"),
					resource.TestCheckResourceAttrSet(dataSourceName, "routes.0.cidr"),
					resource.TestCheckResourceAttrSet(dataSourceName, "routes.0.route_direction"),
					resource.TestCheckResourceAttrSet(dataSourceName, "routes.0.route_installed_at"),
					resource.TestCheckResourceAttr(dataSourceName, "virtual_interface_id", virtualInterfaceID),
				),
			},
			{
				Config: testAccVirtualInterfaceRoutesDataSourceConfig_routeDirection(virtualInterfaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "route_direction", "accepted"),
					resource.TestCheckResourceAttr(dataSourceName, "routes.0.route_direction", "accepted"),
					resource.TestCheckResourceAttr(dataSourceName, "virtual_interface_id", virtualInterfaceID),
				),
			},
		},
	})
}

func testAccVirtualInterfaceRoutesDataSourceConfig_basic(virtualInterfaceID string) string {
	return fmt.Sprintf(`
data "aws_dx_virtual_interface_routes" "test" {
  virtual_interface_id = %[1]q
}
`, virtualInterfaceID)
}

func testAccVirtualInterfaceRoutesDataSourceConfig_routeDirection(virtualInterfaceID string) string {
	return fmt.Sprintf(`
data "aws_dx_virtual_interface_routes" "test" {
  virtual_interface_id = %[1]q
  route_direction      = "accepted"
}
`, virtualInterfaceID)
}
