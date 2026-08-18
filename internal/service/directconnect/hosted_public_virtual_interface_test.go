// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"context"
	"fmt"
	"slices"
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

func TestAccDirectConnectHostedPublicVirtualInterface_basic(t *testing.T) {
	ctx := acctest.Context(t)
	connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")

	var vif awstypes.VirtualInterface
	resourceName := "aws_dx_hosted_public_virtual_interface.test"
	accepterResourceName := "aws_dx_hosted_public_virtual_interface_accepter.test"
	rName := fmt.Sprintf("tf-testacc-public-vif-%s", acctest.RandString(t, 10))
	amazonAddress := "175.45.176.5/28"
	customerAddress := "175.45.176.6/28"
	bgpAsn := acctest.RandIntRange(t, 64512, 65534)
	vlan := acctest.RandIntRange(t, 2049, 4094)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckHostedPublicVirtualInterfaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHostedPublicVirtualInterfaceConfig_basic(connectionID, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedPublicVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_address"),
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
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					// Accepter's attributes:
					resource.TestCheckResourceAttrSet(accepterResourceName, names.AttrARN), // nosemgrep:ci.semgrep.acctest.checks.arn-resourceattrset // TODO: need DX Connection for testing
					resource.TestCheckResourceAttr(accepterResourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrPair(accepterResourceName, "virtual_interface_id", resourceName, names.AttrID),
				),
			},
			{
				Config:            testAccHostedPublicVirtualInterfaceConfig_basic(connectionID, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDirectConnectHostedPublicVirtualInterface_reassociation(t *testing.T) {
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
	resourceName := "aws_dx_hosted_public_virtual_interface.test"
	accepterResourceName := "aws_dx_hosted_public_virtual_interface_accepter.test"
	rName := fmt.Sprintf("tf-testacc-hosted-public-vif-%s", acctest.RandString(t, 9))
	addressOctet := acctest.RandIntRange(t, 1, 254)
	amazonAddress := fmt.Sprintf("175.45.%d.5/28", addressOctet)
	customerAddress := fmt.Sprintf("175.45.%d.6/28", addressOctet)
	routeFilterPrefix1 := fmt.Sprintf("175.45.%d.0/24", addressOctet)
	routeFilterPrefix2 := fmt.Sprintf("210.52.%d.0/24", addressOctet)
	bgpAsn := acctest.RandIntRange(t, 64512, 65534)
	vlan := acctest.RandIntRange(t, 2049, 4094)

	accepterAssociationChecks := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttrPair(accepterResourceName, "virtual_interface_id", resourceName, names.AttrID),
		testAccCheckHostedPublicVirtualInterfaceState(&vif,
			awstypes.VirtualInterfaceStateAvailable,
			awstypes.VirtualInterfaceStateDown,
			awstypes.VirtualInterfaceStateVerifying,
		),
	)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckHostedPublicVirtualInterfaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHostedPublicVirtualInterfaceConfig_reassociation(connectionID, rName, amazonAddress, customerAddress, routeFilterPrefix1, routeFilterPrefix2, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedPublicVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					accepterAssociationChecks,
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
				Config: testAccHostedPublicVirtualInterfaceConfig_reassociation(targetConnectionID, rName, amazonAddress, customerAddress, routeFilterPrefix1, routeFilterPrefix2, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedPublicVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, targetConnectionID),
					accepterAssociationChecks,
					testAccCheckHostedPublicVirtualInterfaceIDAndConnection(resourceName, vifID, targetConnectionID, &vif),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccHostedPublicVirtualInterfaceConfig_reassociation(connectionID, rName, amazonAddress, customerAddress, routeFilterPrefix1, routeFilterPrefix2, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedPublicVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					accepterAssociationChecks,
					testAccCheckHostedPublicVirtualInterfaceIDAndConnection(resourceName, vifID, connectionID, &vif),
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

func TestResourceHostedPublicVirtualInterface_schemaNameNotForceNew(t *testing.T) {
	t.Parallel()

	schema := tfdirectconnect.ResourceHostedPublicVirtualInterface().SchemaFunc()
	for _, name := range []string{names.AttrName, names.AttrConnectionID} {
		if schema[name].ForceNew {
			t.Errorf("%s schema ForceNew = true, want false", name)
		}
	}
}

func TestAccDirectConnectHostedPublicVirtualInterface_name(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")

	var (
		vif   awstypes.VirtualInterface
		vifID string
	)
	resourceName := "aws_dx_hosted_public_virtual_interface.test"
	accepterResourceName := "aws_dx_hosted_public_virtual_interface_accepter.test"
	rName := fmt.Sprintf("tf-testacc-hosted-public-vif-%s", acctest.RandString(t, 9))
	rNameUpdated := fmt.Sprintf("tf-testacc-hosted-public-vif-%s", acctest.RandString(t, 9))
	amazonAddress := "175.45.176.9/28"
	customerAddress := "175.45.176.10/28"
	bgpAsn := acctest.RandIntRange(t, 64512, 65534)
	vlan := acctest.RandIntRange(t, 2049, 4094)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckHostedPublicVirtualInterfaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHostedPublicVirtualInterfaceConfig_name(connectionID, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedPublicVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "amazon_address", amazonAddress),
					resource.TestCheckResourceAttr(resourceName, "customer_address", customerAddress),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_filter_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "210.52.110.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "175.45.180.0/22"),
					resource.TestCheckResourceAttrPair(accepterResourceName, "virtual_interface_id", resourceName, names.AttrID),
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
				Config: testAccHostedPublicVirtualInterfaceConfig_name(connectionID, rNameUpdated, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedPublicVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "amazon_address", amazonAddress),
					resource.TestCheckResourceAttr(resourceName, "customer_address", customerAddress),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "route_filter_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "210.52.110.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "175.45.180.0/22"),
					resource.TestCheckResourceAttrPair(accepterResourceName, "virtual_interface_id", resourceName, names.AttrID),
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

func TestAccDirectConnectHostedPublicVirtualInterface_accepterTags(t *testing.T) {
	ctx := acctest.Context(t)
	connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")

	var vif awstypes.VirtualInterface
	resourceName := "aws_dx_hosted_public_virtual_interface.test"
	accepterResourceName := "aws_dx_hosted_public_virtual_interface_accepter.test"
	rName := fmt.Sprintf("tf-testacc-public-vif-%s", acctest.RandString(t, 10))
	amazonAddress := "175.45.176.7/28"
	customerAddress := "175.45.176.8/28"
	bgpAsn := acctest.RandIntRange(t, 64512, 65534)
	vlan := acctest.RandIntRange(t, 2049, 4094)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckHostedPublicVirtualInterfaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHostedPublicVirtualInterfaceConfig_accepterTags(connectionID, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedPublicVirtualInterfaceExists(ctx, t, resourceName, &vif),
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
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					// Accepter's attributes:
					resource.TestCheckResourceAttrSet(accepterResourceName, names.AttrARN), // nosemgrep:ci.semgrep.acctest.checks.arn-resourceattrset // TODO: need DX Connection for testing
					resource.TestCheckResourceAttr(accepterResourceName, acctest.CtTagsPercent, "3"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key2", "Value2a"),
					resource.TestCheckResourceAttrPair(accepterResourceName, "virtual_interface_id", resourceName, names.AttrID),
				),
			},
			{
				Config: testAccHostedPublicVirtualInterfaceConfig_accepterTagsUpdated(connectionID, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedPublicVirtualInterfaceExists(ctx, t, resourceName, &vif),
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
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					// Accepter's attributes:
					resource.TestCheckResourceAttrSet(accepterResourceName, names.AttrARN), // nosemgrep:ci.semgrep.acctest.checks.arn-resourceattrset // TODO: need DX Connection for testing
					resource.TestCheckResourceAttr(accepterResourceName, acctest.CtTagsPercent, "3"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttrPair(accepterResourceName, "virtual_interface_id", resourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccCheckHostedPublicVirtualInterfaceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckVirtualInterfaceDestroy(ctx, t, s, "aws_dx_hosted_public_virtual_interface")
	}
}

func testAccCheckHostedPublicVirtualInterfaceExists(ctx context.Context, t *testing.T, name string, vif *awstypes.VirtualInterface) resource.TestCheckFunc {
	return testAccCheckVirtualInterfaceExists(ctx, t, name, vif)
}

func testAccCheckHostedPublicVirtualInterfaceState(vif *awstypes.VirtualInterface, want ...awstypes.VirtualInterfaceState) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if slices.Contains(want, vif.VirtualInterfaceState) {
			return nil
		}

		return fmt.Errorf("Virtual Interface state = %s, want one of %v", vif.VirtualInterfaceState, want)
	}
}

func testAccCheckHostedPublicVirtualInterfaceIDAndConnection(resourceName, vifID, connectionID string, vif *awstypes.VirtualInterface) resource.TestCheckFunc {
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

func testAccHostedPublicVirtualInterfaceConfig_base(cid, rName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
# Creator
resource "aws_dx_hosted_public_virtual_interface" "test" {
  address_family   = "ipv4"
  amazon_address   = %[3]q
  bgp_asn          = %[5]d
  connection_id    = %[1]q
  customer_address = %[4]q
  name             = %[2]q
  owner_account_id = data.aws_caller_identity.accepter.account_id
  vlan             = %[6]d

  route_filter_prefixes = [
    "175.45.176.0/22",
    "210.52.109.0/24",
  ]
}

# Accepter
data "aws_caller_identity" "accepter" {
  provider = "awsalternate"
}
`, cid, rName, amzAddr, custAddr, bgpAsn, vlan))
}

func testAccHostedPublicVirtualInterfaceConfig_basic(cid, rName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return acctest.ConfigCompose(testAccHostedPublicVirtualInterfaceConfig_base(cid, rName, amzAddr, custAddr, bgpAsn, vlan), `
resource "aws_dx_hosted_public_virtual_interface_accepter" "test" {
  provider = "awsalternate"

  virtual_interface_id = aws_dx_hosted_public_virtual_interface.test.id
}
`)
}

func testAccHostedPublicVirtualInterfaceConfig_reassociation(cid, rName, amzAddr, custAddr, routeFilterPrefix1, routeFilterPrefix2 string, bgpAsn, vlan int) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
# Allocator
resource "aws_dx_hosted_public_virtual_interface" "test" {
  address_family   = "ipv4"
  amazon_address   = %[3]q
  bgp_asn          = %[7]d
  connection_id    = %[1]q
  customer_address = %[4]q
  name             = %[2]q
  owner_account_id = data.aws_caller_identity.accepter.account_id
  vlan             = %[8]d

  route_filter_prefixes = [
    %[5]q,
    %[6]q,
  ]
}

# Accepter
# The accepter configuration stays unchanged while the allocator reassociates the VIF.
data "aws_caller_identity" "accepter" {
  provider = "awsalternate"
}

resource "aws_dx_hosted_public_virtual_interface_accepter" "test" {
  provider = "awsalternate"

  virtual_interface_id = aws_dx_hosted_public_virtual_interface.test.id
}
`, cid, rName, amzAddr, custAddr, routeFilterPrefix1, routeFilterPrefix2, bgpAsn, vlan))
}

func testAccHostedPublicVirtualInterfaceConfig_name(cid, rName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
resource "aws_dx_hosted_public_virtual_interface" "test" {
  address_family   = "ipv4"
  amazon_address   = %[3]q
  bgp_asn          = %[5]d
  connection_id    = %[1]q
  customer_address = %[4]q
  name             = %[2]q
  owner_account_id = data.aws_caller_identity.accepter.account_id
  vlan             = %[6]d

  route_filter_prefixes = [
    "210.52.110.0/24",
    "175.45.180.0/22",
  ]
}

data "aws_caller_identity" "accepter" {
  provider = "awsalternate"
}

resource "aws_dx_hosted_public_virtual_interface_accepter" "test" {
  provider = "awsalternate"

  virtual_interface_id = aws_dx_hosted_public_virtual_interface.test.id
}
`, cid, rName, amzAddr, custAddr, bgpAsn, vlan))
}

func testAccHostedPublicVirtualInterfaceConfig_accepterTags(cid, rName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return acctest.ConfigCompose(testAccHostedPublicVirtualInterfaceConfig_base(cid, rName, amzAddr, custAddr, bgpAsn, vlan), fmt.Sprintf(`
resource "aws_dx_hosted_public_virtual_interface_accepter" "test" {
  provider = "awsalternate"

  virtual_interface_id = aws_dx_hosted_public_virtual_interface.test.id

  tags = {
    Name = %[1]q
    Key1 = "Value1"
    Key2 = "Value2a"
  }
}
`, rName))
}

func testAccHostedPublicVirtualInterfaceConfig_accepterTagsUpdated(cid, rName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return acctest.ConfigCompose(testAccHostedPublicVirtualInterfaceConfig_base(cid, rName, amzAddr, custAddr, bgpAsn, vlan), fmt.Sprintf(`
resource "aws_dx_hosted_public_virtual_interface_accepter" "test" {
  provider = "awsalternate"

  virtual_interface_id = aws_dx_hosted_public_virtual_interface.test.id

  tags = {
    Name = %[1]q
    Key2 = "Value2b"
    Key3 = "Value3"
  }
}
`, rName))
}
