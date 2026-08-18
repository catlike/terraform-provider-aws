// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectConnectLag_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var lag awstypes.Lag
	resourceName := "aws_dx_lag.test"
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLagDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLagConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLagExists(ctx, t, resourceName, &lag),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "directconnect", regexache.MustCompile(`dxlag/.+`)),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrConnectionID),
					resource.TestCheckResourceAttr(resourceName, "connection_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "child_connection_tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "connections_bandwidth", "10Gbps"),
					resource.TestCheckResourceAttr(resourceName, "number_of_connections", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "has_logical_redundancy"),
					resource.TestCheckResourceAttrSet(resourceName, "jumbo_frame_capable"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLocation),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, names.AttrProviderName, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				Config: testAccLagConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLagExists(ctx, t, resourceName, &lag),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "directconnect", regexache.MustCompile(`dxlag/.+`)),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrConnectionID),
					resource.TestCheckResourceAttr(resourceName, "connection_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "child_connection_tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "connections_bandwidth", "10Gbps"),
					resource.TestCheckResourceAttr(resourceName, "number_of_connections", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "has_logical_redundancy"),
					resource.TestCheckResourceAttrSet(resourceName, "jumbo_frame_capable"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLocation),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, names.AttrProviderName, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccDirectConnectLag_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var lag awstypes.Lag
	resourceName := "aws_dx_lag.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLagDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLagConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLagExists(ctx, t, resourceName, &lag),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdirectconnect.ResourceLag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccDirectConnectLag_connectionID(t *testing.T) {
	ctx := acctest.Context(t)
	var lag awstypes.Lag
	resourceName := "aws_dx_lag.test"
	connectionResourceName := "aws_dx_connection.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLagDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLagConfig_connectionID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLagExists(ctx, t, resourceName, &lag),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "directconnect", regexache.MustCompile(`dxlag/.+`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrConnectionID, connectionResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "connections_bandwidth", "10Gbps"),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "has_logical_redundancy"),
					resource.TestCheckResourceAttrSet(resourceName, "jumbo_frame_capable"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLocation),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, names.AttrProviderName, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrConnectionID, names.AttrForceDestroy},
			},
		},
	})
}

func TestAccDirectConnectLag_providerName(t *testing.T) {
	ctx := acctest.Context(t)
	var lag awstypes.Lag
	resourceName := "aws_dx_lag.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLagDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLagConfig_providerName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLagExists(ctx, t, resourceName, &lag),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "directconnect", regexache.MustCompile(`dxlag/.+`)),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrConnectionID),
					resource.TestCheckResourceAttr(resourceName, "connections_bandwidth", "10Gbps"),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "has_logical_redundancy"),
					resource.TestCheckResourceAttrSet(resourceName, "jumbo_frame_capable"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLocation),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrProviderName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccDirectConnectLag_provisionedConnections(t *testing.T) {
	ctx := acctest.Context(t)
	var lag awstypes.Lag
	var initialConnectionIDs []string
	var replacementConnectionIDs []string
	resourceName := "aws_dx_lag.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLagDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLagConfig_provisionedConnections(rName, 2, acctest.CtValue1, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLagExists(ctx, t, resourceName, &lag),
					testAccCheckLagConnectionCount(ctx, t, resourceName, 2),
					testAccCheckLagConnectionIDs(resourceName, 2, &initialConnectionIDs),
					resource.TestCheckResourceAttr(resourceName, "number_of_connections", "2"),
					resource.TestCheckResourceAttr(resourceName, "connection_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "child_connection_tags.%", "1"),
					testAccCheckLagChildConnectionTags(ctx, t, resourceName, map[string]string{acctest.CtKey1: acctest.CtValue1}),
				),
			},
			{
				Config: testAccLagConfig_provisionedConnections(rName, 1, acctest.CtValue1, false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLagExists(ctx, t, resourceName, &lag),
					testAccCheckLagConnectionCount(ctx, t, resourceName, 1),
					testAccCheckLagConnectionsDeleted(ctx, t, &initialConnectionIDs),
					testAccCheckLagConnectionIDs(resourceName, 1, &replacementConnectionIDs),
					resource.TestCheckResourceAttr(resourceName, "number_of_connections", "1"),
					resource.TestCheckResourceAttr(resourceName, "connection_ids.#", "1"),
					testAccCheckLagChildConnectionTags(ctx, t, resourceName, map[string]string{acctest.CtKey1: acctest.CtValue1}),
				),
			},
			{
				Config: testAccLagConfig_provisionedConnections(rName, 1, acctest.CtValue2, false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLagExists(ctx, t, resourceName, &lag),
					testAccCheckLagConnectionCount(ctx, t, resourceName, 1),
					testAccCheckLagConnectionsDeleted(ctx, t, &replacementConnectionIDs),
					resource.TestCheckResourceAttr(resourceName, "number_of_connections", "1"),
					resource.TestCheckResourceAttr(resourceName, "connection_ids.#", "1"),
					testAccCheckLagChildConnectionTags(ctx, t, resourceName, map[string]string{acctest.CtKey1: acctest.CtValue2}),
				),
			},
			{
				Config: testAccLagConfig_provisionedConnections(rName, 1, acctest.CtValue2, true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLagExists(ctx, t, resourceName, &lag),
					testAccCheckLagConnectionCount(ctx, t, resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "connection_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDestroy, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccDirectConnectLag_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var lag awstypes.Lag
	resourceName := "aws_dx_lag.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLagDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLagConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLagExists(ctx, t, resourceName, &lag),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccLagConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLagExists(ctx, t, resourceName, &lag),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccLagConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLagExists(ctx, t, resourceName, &lag),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckLagDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DirectConnectClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dx_lag" {
				continue
			}

			_, err := tfdirectconnect.FindLagByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Direct Connect LAG %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLagExists(ctx context.Context, t *testing.T, name string, v *awstypes.Lag) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.ProviderMeta(ctx, t).DirectConnectClient(ctx)

		output, err := tfdirectconnect.FindLagByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckLagConnectionCount(ctx context.Context, t *testing.T, name string, expectedCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.ProviderMeta(ctx, t).DirectConnectClient(ctx)
		lag, err := tfdirectconnect.FindLagByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if actualCount := len(lag.Connections); actualCount != expectedCount {
			return fmt.Errorf("Direct Connect LAG %s has %d connections, want %d", rs.Primary.ID, actualCount, expectedCount)
		}

		return nil
	}
}

func testAccCheckLagConnectionIDs(name string, expectedCount int, connectionIDs *[]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if connectionIDs == nil {
			return fmt.Errorf("connection ID destination cannot be nil")
		}

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		ids := make([]string, 0, expectedCount)
		seen := make(map[string]struct{}, expectedCount)
		for key, value := range rs.Primary.Attributes {
			if key == "connection_ids.#" || len(key) < len("connection_ids.") || key[:len("connection_ids.")] != "connection_ids." {
				continue
			}
			if value == "" {
				return fmt.Errorf("Direct Connect LAG %s has an empty managed connection ID", rs.Primary.ID)
			}
			if _, ok := seen[value]; ok {
				return fmt.Errorf("Direct Connect LAG %s has duplicate managed connection ID %s", rs.Primary.ID, value)
			}

			seen[value] = struct{}{}
			ids = append(ids, value)
		}
		if len(ids) != expectedCount {
			return fmt.Errorf("Direct Connect LAG %s has %d managed connection IDs in state, want %d", rs.Primary.ID, len(ids), expectedCount)
		}

		*connectionIDs = ids
		return nil
	}
}

func testAccCheckLagConnectionsDeleted(ctx context.Context, t *testing.T, connectionIDs *[]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(*connectionIDs) == 0 {
			return fmt.Errorf("expected at least one managed connection ID to verify deletion")
		}

		conn := acctest.ProviderMeta(ctx, t).DirectConnectClient(ctx)
		for _, connectionID := range *connectionIDs {
			if connectionID == "" {
				return fmt.Errorf("cannot verify deletion of an empty managed connection ID")
			}

			_, err := tfdirectconnect.FindConnectionByID(ctx, conn, connectionID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Direct Connect Connection %s still exists", connectionID)
		}

		return nil
	}
}

func testAccCheckLagChildConnectionTags(ctx context.Context, t *testing.T, name string, expected map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.ProviderMeta(ctx, t).DirectConnectClient(ctx)
		lag, err := tfdirectconnect.FindLagByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		if len(lag.Connections) == 0 {
			return fmt.Errorf("Direct Connect LAG %s has no child connections to verify tags", rs.Primary.ID)
		}

		for _, childConnection := range lag.Connections {
			output, err := conn.DescribeConnections(ctx, &directconnect.DescribeConnectionsInput{
				ConnectionId: childConnection.ConnectionId,
			})
			if err != nil {
				return err
			}
			if len(output.Connections) != 1 {
				return fmt.Errorf("expected one Direct Connect child connection for %s, got %d", aws.ToString(childConnection.ConnectionId), len(output.Connections))
			}

			actual := make(map[string]string, len(output.Connections[0].Tags))
			for _, tag := range output.Connections[0].Tags {
				actual[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
			}
			for key, value := range expected {
				if actual[key] != value {
					return fmt.Errorf("Direct Connect child connection %s tag %q = %q, want %q", aws.ToString(childConnection.ConnectionId), key, actual[key], value)
				}
			}
		}

		return nil
	}
}

func testAccLagConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

resource "aws_dx_lag" "test" {
  name                  = %[1]q
  connections_bandwidth = "10Gbps"
  location              = tolist(data.aws_dx_locations.test.location_codes)[0]
}
`, rName)
}

func testAccLagConfig_provisionedConnections(rName string, numberOfConnections int, childTagValue string, forceDestroy bool) string {
	return fmt.Sprintf(`
resource "aws_dx_lag" "test" {
  name                  = %[1]q
  connections_bandwidth = "10Gbps"
  force_destroy         = %[2]t
  location              = "EqDC2"
  number_of_connections = %[3]d

  child_connection_tags = {
    %[4]q = %[5]q
  }
}
`, rName, forceDestroy, numberOfConnections, acctest.CtKey1, childTagValue)
}

func testAccLagConfig_connectionID(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

resource "aws_dx_lag" "test" {
  name                  = %[1]q
  connection_id         = aws_dx_connection.test.id
  connections_bandwidth = aws_dx_connection.test.bandwidth
  location              = aws_dx_connection.test.location
}

resource "aws_dx_connection" "test" {
  name      = %[1]q
  bandwidth = "10Gbps"
  location  = tolist(data.aws_dx_locations.test.location_codes)[1]
}
`, rName)
}

func testAccLagConfig_providerName(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

data "aws_dx_location" "test" {
  location_code = tolist(data.aws_dx_locations.test.location_codes)[0]
}

resource "aws_dx_lag" "test" {
  name                  = %[1]q
  connections_bandwidth = "10Gbps"
  location              = data.aws_dx_location.test.location_code

  provider_name = data.aws_dx_location.test.available_providers[0]
}
`, rName)
}

func testAccLagConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

resource "aws_dx_lag" "test" {
  name                  = %[1]q
  connections_bandwidth = "10Gbps"
  location              = tolist(data.aws_dx_locations.test.location_codes)[0]
  force_destroy         = true

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccLagConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

resource "aws_dx_lag" "test" {
  name                  = %[1]q
  connections_bandwidth = "10Gbps"
  location              = tolist(data.aws_dx_locations.test.location_codes)[0]
  force_destroy         = true

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
