// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCNetworkInterfacePermission_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var networkInterfacePermission types.NetworkInterfacePermission
	resourceName := "aws_network_interface_permission.test"
	eniResourceName := "aws_network_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCNetworkInterfacePermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfacePermissionConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkInterfacePermissionExists(ctx, resourceName, &networkInterfacePermission),
					resource.TestCheckResourceAttrPair(resourceName, "network_interface_id", eniResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "permission", "INSTANCE-ATTACH"),
					resource.TestCheckResourceAttrSet(resourceName, "aws_account_id"),
					resource.TestCheckResourceAttrSet(resourceName, "permission_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckVPCNetworkInterfacePermissionExists(ctx context.Context, n string, v *types.NetworkInterfacePermission) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Network Interface Permission ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.findNetworkInterfacePermissionByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckVPCNetworkInterfacePermissionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_network_interface_permission" {
				continue
			}

			_, err := tfec2.findNetworkInterfacePermissionByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Network Interface Permission %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccVPCNetworkInterfacePermissionConfig_basic() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-network-interface-permission-basic"
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-network-interface-permission-basic"
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = "tf-acc-network-interface-permission-basic"
  }
}

data "aws_caller_identity" "current" {}

resource "aws_network_interface_permission" "test" {
  network_interface_id = aws_network_interface.test.id
  aws_account_id      = data.aws_caller_identity.current.account_id
  permission          = "INSTANCE-ATTACH"
}
`)
}
