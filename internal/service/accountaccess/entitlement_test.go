// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfaccountaccess "github.com/hashicorp/terraform-provider-aws/internal/service/accountaccess"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// configVarApplicationArn / Principal / Role read the values needed by the
// entitlement_basic.gtpl-generated configs from environment variables. The
// test will skip cleanly if any are missing — acceptance tests can't run
// without an allowlisted AWS account that has a real Application + IdC user
// or group + a target IAM role.
const (
	envVarApplicationArn = "TF_VAR_application_arn"
	envVarPrincipalID    = "TF_VAR_principal_id"
	envVarPrincipalType  = "TF_VAR_principal_type"
	envVarRoleARN        = "TF_VAR_role_arn"
)

func TestAccAccountAccessEntitlement_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	acctest.SkipIfEnvVarNotSet(t, envVarApplicationArn)
	acctest.SkipIfEnvVarNotSet(t, envVarPrincipalID)
	acctest.SkipIfEnvVarNotSet(t, envVarRoleARN)

	var v accountaccess.GetEntitlementOutput
	resourceName := "aws_accountaccess_entitlement.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountAccessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntitlementDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Entitlement/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEntitlementExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "entitlement_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAccountID),
				),
			},
		},
	})
}

func TestAccAccountAccessEntitlement_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	acctest.SkipIfEnvVarNotSet(t, envVarApplicationArn)
	acctest.SkipIfEnvVarNotSet(t, envVarPrincipalID)
	acctest.SkipIfEnvVarNotSet(t, envVarRoleARN)

	var v accountaccess.GetEntitlementOutput
	resourceName := "aws_accountaccess_entitlement.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountAccessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntitlementDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Entitlement/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEntitlementExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfaccountaccess.ResourceEntitlement, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEntitlementDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AccountAccessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_accountaccess_entitlement" {
				continue
			}

			applicationArn := rs.Primary.Attributes["application_arn"]
			entitlementID := rs.Primary.Attributes["entitlement_id"]

			_, err := tfaccountaccess.FindEntitlementByTwoPartKey(ctx, conn, applicationArn, entitlementID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Account Access Entitlement %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckEntitlementExists(ctx context.Context, t *testing.T, n string, v *accountaccess.GetEntitlementOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return errors.New("No Account Access Entitlement ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).AccountAccessClient(ctx)

		applicationArn := rs.Primary.Attributes["application_arn"]
		entitlementID := rs.Primary.Attributes["entitlement_id"]

		output, err := tfaccountaccess.FindEntitlementByTwoPartKey(ctx, conn, applicationArn, entitlementID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}
