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

func TestAccDirectConnectPublicVirtualInterface_basic(t *testing.T) {
	ctx := acctest.Context(t)
	connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")

	var vif awstypes.VirtualInterface
	resourceName := "aws_dx_public_virtual_interface.test"
	rName := fmt.Sprintf("tf-testacc-public-vif-%s", acctest.RandString(t, 10))
	// DirectConnectClientException: Amazon Address is not allowed to contain a private IP
	// DirectConnectClientException: Amazon Address and Customer Address must be in the same CIDR
	// DirectConnectClientException: Amazon Address is address 0 on its subnet.
	// DirectConnectClientException: Amazon Address is the broadcast address on its subnet.
	amazonAddress := "175.45.176.1/28"
	customerAddress := "175.45.176.2/28"
	bgpAsn := acctest.RandIntRange(t, 64512, 65534)
	vlan := acctest.RandIntRange(t, 2049, 4094)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublicVirtualInterfaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPublicVirtualInterfaceConfig_basic(connectionID, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "amazon_address", amazonAddress),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.ToString(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					resource.TestCheckResourceAttr(resourceName, "customer_address", customerAddress),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_filter_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "210.52.109.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "175.45.176.0/22"),
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

func TestResourcePublicVirtualInterface_schemaNameAndConnectionIDNotForceNew(t *testing.T) {
	t.Parallel()

	schema := tfdirectconnect.ResourcePublicVirtualInterface().SchemaFunc()

	for _, attributeName := range []string{names.AttrName, names.AttrConnectionID} {
		if schema[attributeName].ForceNew {
			t.Errorf("%s schema ForceNew = true, want false", attributeName)
		}
	}
}

func TestAccDirectConnectPublicVirtualInterface_name(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")

	var (
		vif   awstypes.VirtualInterface
		vifID string
	)
	resourceName := "aws_dx_public_virtual_interface.test"
	rName := fmt.Sprintf("tf-testacc-public-vif-%s", acctest.RandString(t, 10))
	rNameUpdated := fmt.Sprintf("tf-testacc-public-vif-%s", acctest.RandString(t, 10))
	amazonAddress := "216.3.128.1/30"
	customerAddress := "216.3.128.2/30"
	bgpAsn := acctest.RandIntRange(t, 64512, 65534)
	vlan := acctest.RandIntRange(t, 2049, 4094)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublicVirtualInterfaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPublicVirtualInterfaceConfig_name(connectionID, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicVirtualInterfaceExists(ctx, t, resourceName, &vif),
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
				Config: testAccPublicVirtualInterfaceConfig_name(connectionID, rNameUpdated, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicVirtualInterfaceExists(ctx, t, resourceName, &vif),
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

func TestAccDirectConnectPublicVirtualInterface_reassociation(t *testing.T) {
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
	resourceName := "aws_dx_public_virtual_interface.test"
	rName := fmt.Sprintf("tf-testacc-public-vif-%s", acctest.RandString(t, 9))
	amazonAddress := "146.112.128.1/30"
	customerAddress := "146.112.128.2/30"
	bgpAsn := acctest.RandIntRange(t, 64512, 65534)
	vlan := acctest.RandIntRange(t, 2049, 4094)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublicVirtualInterfaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPublicVirtualInterfaceConfig_reassociation(connectionID, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicVirtualInterfaceExists(ctx, t, resourceName, &vif),
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
				Config: testAccPublicVirtualInterfaceConfig_reassociation(targetConnectionID, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, targetConnectionID),
					testAccCheckPublicVirtualInterfaceIDAndConnection(resourceName, vifID, targetConnectionID, &vif),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccPublicVirtualInterfaceConfig_reassociation(connectionID, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					testAccCheckPublicVirtualInterfaceIDAndConnection(resourceName, vifID, connectionID, &vif),
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

func TestAccDirectConnectPublicVirtualInterface_tags(t *testing.T) {
	ctx := acctest.Context(t)
	connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")

	var vif awstypes.VirtualInterface
	resourceName := "aws_dx_public_virtual_interface.test"
	rName := fmt.Sprintf("tf-testacc-public-vif-%s", acctest.RandString(t, 10))
	amazonAddress := "175.45.176.3/28"
	customerAddress := "175.45.176.4/28"
	bgpAsn := acctest.RandIntRange(t, 64512, 65534)
	vlan := acctest.RandIntRange(t, 2049, 4094)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublicVirtualInterfaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPublicVirtualInterfaceConfig_tags(connectionID, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "amazon_address", amazonAddress),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.ToString(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					resource.TestCheckResourceAttr(resourceName, "customer_address", customerAddress),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_filter_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "210.52.109.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "175.45.176.0/22"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
				),
			},
			{
				Config: testAccPublicVirtualInterfaceConfig_tagsUpdated(connectionID, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "amazon_address", amazonAddress),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.ToString(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					resource.TestCheckResourceAttr(resourceName, "customer_address", customerAddress),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_filter_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "210.52.109.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "175.45.176.0/22"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
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

func testAccCheckPublicVirtualInterfaceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckVirtualInterfaceDestroy(ctx, t, s, "aws_dx_public_virtual_interface")
	}
}

func testAccCheckPublicVirtualInterfaceExists(ctx context.Context, t *testing.T, name string, vif *awstypes.VirtualInterface) resource.TestCheckFunc {
	return testAccCheckVirtualInterfaceExists(ctx, t, name, vif)
}

func testAccCheckPublicVirtualInterfaceIDAndConnection(resourceName, vifID, connectionID string, vif *awstypes.VirtualInterface) resource.TestCheckFunc {
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

func testAccPublicVirtualInterfaceConfig_basic(cid, rName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return fmt.Sprintf(`
resource "aws_dx_public_virtual_interface" "test" {
  address_family   = "ipv4"
  amazon_address   = %[3]q
  bgp_asn          = %[5]d
  connection_id    = %[1]q
  customer_address = %[4]q
  name             = %[2]q
  vlan             = %[6]d

  route_filter_prefixes = [
    "175.45.176.0/22",
    "210.52.109.0/24",
  ]
}
`, cid, rName, amzAddr, custAddr, bgpAsn, vlan)
}

func testAccPublicVirtualInterfaceConfig_reassociation(cid, vifName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return fmt.Sprintf(`
resource "aws_dx_public_virtual_interface" "test" {
  address_family   = "ipv4"
  amazon_address   = %[3]q
  bgp_asn          = %[5]d
  connection_id    = %[1]q
  customer_address = %[4]q
  name             = %[2]q
  vlan             = %[6]d

  route_filter_prefixes = [
    "146.112.128.0/24",
    "209.165.200.0/24",
  ]
}
`, cid, vifName, amzAddr, custAddr, bgpAsn, vlan)
}

func testAccPublicVirtualInterfaceConfig_name(cid, vifName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return fmt.Sprintf(`
resource "aws_dx_public_virtual_interface" "test" {
  address_family   = "ipv4"
  amazon_address   = %[3]q
  bgp_asn          = %[5]d
  connection_id    = %[1]q
  customer_address = %[4]q
  name             = %[2]q
  vlan             = %[6]d

  route_filter_prefixes = [
    "216.3.128.0/24",
    "210.52.110.0/24",
  ]
}
`, cid, vifName, amzAddr, custAddr, bgpAsn, vlan)
}

func testAccPublicVirtualInterfaceConfig_tags(cid, rName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return fmt.Sprintf(`
resource "aws_dx_public_virtual_interface" "test" {
  address_family   = "ipv4"
  amazon_address   = %[3]q
  bgp_asn          = %[5]d
  connection_id    = %[1]q
  customer_address = %[4]q
  name             = %[2]q
  vlan             = %[6]d

  route_filter_prefixes = [
    "175.45.176.0/22",
    "210.52.109.0/24",
  ]

  tags = {
    Name = %[2]q
    Key1 = "Value1"
    Key2 = "Value2a"
  }
}
`, cid, rName, amzAddr, custAddr, bgpAsn, vlan)
}

func testAccPublicVirtualInterfaceConfig_tagsUpdated(cid, rName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return fmt.Sprintf(`
resource "aws_dx_public_virtual_interface" "test" {
  address_family   = "ipv4"
  amazon_address   = %[3]q
  bgp_asn          = %[5]d
  connection_id    = %[1]q
  customer_address = %[4]q
  name             = %[2]q
  vlan             = %[6]d

  route_filter_prefixes = [
    "175.45.176.0/22",
    "210.52.109.0/24",
  ]

  tags = {
    Name = %[2]q
    Key2 = "Value2b"
    Key3 = "Value3"
  }
}
`, cid, rName, amzAddr, custAddr, bgpAsn, vlan)
}
