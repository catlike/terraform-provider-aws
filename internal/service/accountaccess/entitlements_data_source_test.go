// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAccountAccessEntitlementsDataSource_byPrincipal(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	acctest.SkipIfEnvVarNotSet(t, envVarApplicationArn)
	acctest.SkipIfEnvVarNotSet(t, envVarPrincipalID)
	acctest.SkipIfEnvVarNotSet(t, envVarRoleARN)

	dataSourceName := "data.aws_accountaccess_entitlements.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountAccessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEntitlementsDataSourceConfig_byPrincipal(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// At least one entitlement is returned (the one we just created).
					resource.TestCheckResourceAttrSet(dataSourceName, "entitlements.#"),
				),
			},
		},
	})
}

func testAccEntitlementsDataSourceConfig_byPrincipal() string {
	return `
variable "application_arn" {
  type     = string
  nullable = false
}

variable "principal_id" {
  type     = string
  nullable = false
}

variable "principal_type" {
  type    = string
  default = "USER"
}

variable "role_arn" {
  type     = string
  nullable = false
}

resource "aws_accountaccess_entitlement" "test" {
  application_arn = var.application_arn
  principal_id    = var.principal_id
  principal_type  = var.principal_type
  role_arn        = var.role_arn
}

data "aws_accountaccess_entitlements" "test" {
  application_arn = aws_accountaccess_entitlement.test.application_arn
  principal_id    = aws_accountaccess_entitlement.test.principal_id
  principal_type  = aws_accountaccess_entitlement.test.principal_type
}
`
}
