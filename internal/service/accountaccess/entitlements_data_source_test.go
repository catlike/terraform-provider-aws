// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccountAccessEntitlementsDataSource_byPrincipal(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_accountaccess_entitlements.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckRegion(t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountAccessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntitlementDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEntitlementsDataSourceConfig_byPrincipal(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("entitlements"), knownvalue.ListSizeExact(1)),
				},
			},
		},
	})
}

func testAccAccountAccessEntitlementsDataSource_byRole(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_accountaccess_entitlements.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckRegion(t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountAccessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntitlementDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEntitlementsDataSourceConfig_byRole(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("entitlements"), knownvalue.ListSizeExact(1)),
				},
			},
		},
	})
}

func testAccEntitlementsDataSourceConfig_byPrincipal(rName string) string {
	return acctest.ConfigCompose(testAccEntitlementConfig_user(rName), `
data "aws_accountaccess_entitlements" "test" {
  application_arn = aws_accountaccess_entitlement.test.application_arn
  principal_id    = aws_accountaccess_entitlement.test.principal_id
  principal_type  = aws_accountaccess_entitlement.test.principal_type
}
`)
}

func testAccEntitlementsDataSourceConfig_byRole(rName string) string {
	return acctest.ConfigCompose(testAccEntitlementConfig_user(rName), `
data "aws_accountaccess_entitlements" "test" {
  application_arn = aws_accountaccess_entitlement.test.application_arn
  role_arn        = aws_accountaccess_entitlement.test.role_arn
}
`)
}
