// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectConnectPrivateVirtualInterface_basic(t *testing.T) {
	ctx := acctest.Context(t)
	connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")

	var vif awstypes.VirtualInterface
	resourceName := "aws_dx_private_virtual_interface.test"
	vpnGatewayResourceName := "aws_vpn_gateway.test"
	rName := fmt.Sprintf("tf-testacc-private-vif-%s", acctest.RandString(t, 9))
	bgpAsn := acctest.RandIntRange(t, 64512, 65534)
	vlan := acctest.RandIntRange(t, 2049, 4094)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateVirtualInterfaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateVirtualInterfaceConfig_basic(connectionID, rName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_address"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.ToString(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					resource.TestCheckResourceAttrSet(resourceName, "customer_address"),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_gateway_id", vpnGatewayResourceName, names.AttrID),
				),
			},
			{
				Config: testAccPrivateVirtualInterfaceConfig_updated(connectionID, rName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_address"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.ToString(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					resource.TestCheckResourceAttrSet(resourceName, "customer_address"),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mtu", "9001"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_gateway_id", vpnGatewayResourceName, names.AttrID),
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

func TestResourcePrivateVirtualInterface_schemaNameAndConnectionIDNotForceNew(t *testing.T) {
	t.Parallel()

	schema := tfdirectconnect.ResourcePrivateVirtualInterface().SchemaFunc()

	for _, attributeName := range []string{names.AttrName, names.AttrConnectionID} {
		if schema[attributeName].ForceNew {
			t.Errorf("%s schema ForceNew = true, want false", attributeName)
		}
	}
}

func TestAccDirectConnectPrivateVirtualInterface_name(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")

	var (
		vif   awstypes.VirtualInterface
		vifID string
	)
	resourceName := "aws_dx_private_virtual_interface.test"
	rName := fmt.Sprintf("tf-testacc-private-vif-%s", acctest.RandString(t, 9))
	rNameUpdated := fmt.Sprintf("tf-testacc-private-vif-%s", acctest.RandString(t, 9))
	bgpAsn := acctest.RandIntRange(t, 64512, 65534)
	vlan := acctest.RandIntRange(t, 2049, 4094)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateVirtualInterfaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateVirtualInterfaceConfig_name(connectionID, rName, rName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateVirtualInterfaceExists(ctx, t, resourceName, &vif),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("Not found: %s", resourceName)
						}

						vifID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: testAccPrivateVirtualInterfaceConfig_name(connectionID, rName, rNameUpdated, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("Not found: %s", resourceName)
						}

						if rs.Primary.ID != vifID {
							return fmt.Errorf("Virtual Interface ID = %s, want %s", rs.Primary.ID, vifID)
						}

						return nil
					},
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDirectConnectPrivateVirtualInterface_reassociation(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")
	targetConnectionID := acctest.SkipIfEnvVarNotSet(t, "DX_TARGET_CONNECTION_ID")
	if connectionID == targetConnectionID {
		t.Fatal("DX_CONNECTION_ID and DX_TARGET_CONNECTION_ID must differ")
	}

	var (
		vif   awstypes.VirtualInterface
		vifID string
	)
	resourceName := "aws_dx_private_virtual_interface.test"
	rName := fmt.Sprintf("tf-testacc-private-vif-%s", acctest.RandString(t, 9))
	amzAsn := acctest.RandIntRange(t, 64512, 65534)
	bgpAsn := acctest.RandIntRange(t, 64512, 65534)
	vlan := acctest.RandIntRange(t, 2049, 4094)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateVirtualInterfaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateVirtualInterfaceConfig_gateway(connectionID, rName, amzAsn, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("Not found: %s", resourceName)
						}

						vifID = rs.Primary.ID
						if aws.ToString(vif.ConnectionId) != connectionID {
							return fmt.Errorf("remote connection ID = %s, want %s", aws.ToString(vif.ConnectionId), connectionID)
						}
						return nil
					},
				),
			},
			{
				Config: testAccPrivateVirtualInterfaceConfig_gateway(targetConnectionID, rName, amzAsn, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, targetConnectionID),
					testAccCheckPrivateVirtualInterfaceIDAndConnection(resourceName, vifID, targetConnectionID, &vif),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccPrivateVirtualInterfaceConfig_gateway(connectionID, rName, amzAsn, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					testAccCheckPrivateVirtualInterfaceIDAndConnection(resourceName, vifID, connectionID, &vif),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDirectConnectPrivateVirtualInterface_tags(t *testing.T) {
	ctx := acctest.Context(t)
	connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")

	var vif awstypes.VirtualInterface
	resourceName := "aws_dx_private_virtual_interface.test"
	vpnGatewayResourceName := "aws_vpn_gateway.test"
	rName := fmt.Sprintf("tf-testacc-private-vif-%s", acctest.RandString(t, 9))
	bgpAsn := acctest.RandIntRange(t, 64512, 65534)
	vlan := acctest.RandIntRange(t, 2049, 4094)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateVirtualInterfaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateVirtualInterfaceConfig_tags(connectionID, rName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_address"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.ToString(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					resource.TestCheckResourceAttrSet(resourceName, "customer_address"),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_gateway_id", vpnGatewayResourceName, names.AttrID),
				),
			},
			{
				Config: testAccPrivateVirtualInterfaceConfig_tagsUpdated(connectionID, rName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_address"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.ToString(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					resource.TestCheckResourceAttrSet(resourceName, "customer_address"),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_gateway_id", vpnGatewayResourceName, names.AttrID),
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

func TestAccDirectConnectPrivateVirtualInterface_dxGateway(t *testing.T) {
	ctx := acctest.Context(t)
	connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")

	var vif awstypes.VirtualInterface
	resourceName := "aws_dx_private_virtual_interface.test"
	dxGatewayResourceName := "aws_dx_gateway.test"
	rName := fmt.Sprintf("tf-testacc-private-vif-%s", acctest.RandString(t, 9))
	amzAsn := acctest.RandIntRange(t, 64512, 65534)
	bgpAsn := acctest.RandIntRange(t, 64512, 65534)
	vlan := acctest.RandIntRange(t, 2049, 4094)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateVirtualInterfaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateVirtualInterfaceConfig_gateway(connectionID, rName, amzAsn, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_address"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.ToString(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					resource.TestCheckResourceAttrSet(resourceName, "customer_address"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", dxGatewayResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
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

func TestAccDirectConnectPrivateVirtualInterface_siteLink(t *testing.T) {
	ctx := acctest.Context(t)
	connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")

	var vif awstypes.VirtualInterface
	resourceName := "aws_dx_private_virtual_interface.test"
	dxGatewayResourceName := "aws_dx_gateway.test"
	rName := fmt.Sprintf("tf-testacc-private-vif-%s", acctest.RandString(t, 9))
	amzAsn := acctest.RandIntRange(t, 64512, 65534)
	bgpAsn := acctest.RandIntRange(t, 64512, 65534)
	vlan := acctest.RandIntRange(t, 2049, 4094)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateVirtualInterfaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateVirtualInterfaceConfig_siteLinkBasic(connectionID, rName, amzAsn, bgpAsn, vlan, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_address"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.ToString(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					resource.TestCheckResourceAttrSet(resourceName, "customer_address"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", dxGatewayResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					resource.TestCheckResourceAttr(resourceName, "sitelink_enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccPrivateVirtualInterfaceConfig_siteLinkUpdated(connectionID, rName, amzAsn, bgpAsn, vlan, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_address"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.ToString(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					resource.TestCheckResourceAttrSet(resourceName, "customer_address"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", dxGatewayResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					resource.TestCheckResourceAttr(resourceName, "sitelink_enabled", acctest.CtFalse),
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

func testAccCheckPrivateVirtualInterfaceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckVirtualInterfaceDestroy(ctx, t, s, "aws_dx_private_virtual_interface")
	}
}

func testAccCheckPrivateVirtualInterfaceExists(ctx context.Context, t *testing.T, name string, vif *awstypes.VirtualInterface) resource.TestCheckFunc {
	return testAccCheckVirtualInterfaceExists(ctx, t, name, vif)
}

func testAccCheckPrivateVirtualInterfaceIDAndConnection(resourceName, vifID, connectionID string, vif *awstypes.VirtualInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID != vifID {
			return fmt.Errorf("Virtual Interface ID = %s, want %s", rs.Primary.ID, vifID)
		}

		if aws.ToString(vif.ConnectionId) != connectionID {
			return fmt.Errorf("remote connection ID = %s, want %s", aws.ToString(vif.ConnectionId), connectionID)
		}

		return nil
	}
}

func testAccPrivateVirtualInterfaceConfig_vpnGateway(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}`, rName)
}

func testAccPrivateVirtualInterfaceConfig_basic(cid, rName string, bgpAsn, vlan int) string {
	return acctest.ConfigCompose(testAccPrivateVirtualInterfaceConfig_vpnGateway(rName), fmt.Sprintf(`
resource "aws_dx_private_virtual_interface" "test" {
  address_family = "ipv4"
  bgp_asn        = %[3]d
  connection_id  = %[1]q
  name           = %[2]q
  vlan           = %[4]d
  vpn_gateway_id = aws_vpn_gateway.test.id
}
`, cid, rName, bgpAsn, vlan))
}

func testAccPrivateVirtualInterfaceConfig_name(cid, rName, vifName string, bgpAsn, vlan int) string {
	return acctest.ConfigCompose(testAccPrivateVirtualInterfaceConfig_vpnGateway(rName), fmt.Sprintf(`
resource "aws_dx_private_virtual_interface" "test" {
  address_family = "ipv4"
  bgp_asn        = %[4]d
  connection_id  = %[1]q
  name           = %[3]q
  vlan           = %[5]d
  vpn_gateway_id = aws_vpn_gateway.test.id
}
`, cid, rName, vifName, bgpAsn, vlan))
}

func testAccPrivateVirtualInterfaceConfig_updated(cid, rName string, bgpAsn, vlan int) string {
	return acctest.ConfigCompose(testAccPrivateVirtualInterfaceConfig_vpnGateway(rName), fmt.Sprintf(`
resource "aws_dx_private_virtual_interface" "test" {
  address_family = "ipv4"
  bgp_asn        = %[3]d
  connection_id  = %[1]q
  mtu            = 9001
  name           = %[2]q
  vlan           = %[4]d
  vpn_gateway_id = aws_vpn_gateway.test.id
}
`, cid, rName, bgpAsn, vlan))
}

func testAccPrivateVirtualInterfaceConfig_tags(cid, rName string, bgpAsn, vlan int) string {
	return acctest.ConfigCompose(testAccPrivateVirtualInterfaceConfig_vpnGateway(rName), fmt.Sprintf(`
resource "aws_dx_private_virtual_interface" "test" {
  address_family = "ipv4"
  bgp_asn        = %[3]d
  connection_id  = %[1]q
  name           = %[2]q
  vlan           = %[4]d
  vpn_gateway_id = aws_vpn_gateway.test.id

  tags = {
    Name = %[2]q
    Key1 = "Value1"
    Key2 = "Value2a"
  }
}
`, cid, rName, bgpAsn, vlan))
}

func testAccPrivateVirtualInterfaceConfig_tagsUpdated(cid, rName string, bgpAsn, vlan int) string {
	return acctest.ConfigCompose(testAccPrivateVirtualInterfaceConfig_vpnGateway(rName), fmt.Sprintf(`
resource "aws_dx_private_virtual_interface" "test" {
  address_family = "ipv4"
  bgp_asn        = %[3]d
  connection_id  = %[1]q
  name           = %[2]q
  vlan           = %[4]d
  vpn_gateway_id = aws_vpn_gateway.test.id

  tags = {
    Name = %[2]q
    Key2 = "Value2b"
    Key3 = "Value3"
  }
}
`, cid, rName, bgpAsn, vlan))
}

func testAccPrivateVirtualInterfaceConfig_gateway(cid, rName string, amzAsn, bgpAsn, vlan int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  amazon_side_asn = %[3]d
  name            = %[2]q
}

resource "aws_dx_private_virtual_interface" "test" {
  address_family = "ipv4"
  bgp_asn        = %[4]d
  dx_gateway_id  = aws_dx_gateway.test.id
  connection_id  = %[1]q
  name           = %[2]q
  vlan           = %[5]d
}
`, cid, rName, amzAsn, bgpAsn, vlan)
}

func testAccPrivateVirtualInterfaceConfig_siteLinkBasic(cid, rName string, amzAsn, bgpAsn, vlan int, sitelink_enabled bool) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  amazon_side_asn = %[3]d
  name            = %[2]q
}

resource "aws_dx_private_virtual_interface" "test" {
  address_family   = "ipv4"
  bgp_asn          = %[4]d
  dx_gateway_id    = aws_dx_gateway.test.id
  connection_id    = %[1]q
  name             = %[2]q
  vlan             = %[5]d
  sitelink_enabled = %[6]t
}
`, cid, rName, amzAsn, bgpAsn, vlan, sitelink_enabled)
}

func testAccPrivateVirtualInterfaceConfig_siteLinkUpdated(cid, rName string, amzAsn, bgpAsn, vlan int, sitelink_enabled bool) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  amazon_side_asn = %[3]d
  name            = %[2]q
}

resource "aws_dx_private_virtual_interface" "test" {
  address_family   = "ipv4"
  bgp_asn          = %[4]d
  dx_gateway_id    = aws_dx_gateway.test.id
  connection_id    = %[1]q
  name             = %[2]q
  vlan             = %[5]d
  sitelink_enabled = %[6]t
}
`, cid, rName, amzAsn, bgpAsn, vlan, sitelink_enabled)
}
