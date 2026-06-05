// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAccountAccessApplicationDataSource_byInstance(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	acctest.SkipIfEnvVarNotSet(t, envVarApplicationArn) // proxy: same env-var bundle indicates a configured test account

	dataSourceName := "data.aws_accountaccess_application.test"
	resourceName := "aws_accountaccess_application.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountAccessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationDataSourceConfig_byInstance(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "identity_center_instance_arn", resourceName, "identity_center_instance_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, resourceName, names.AttrStatus),
				),
			},
		},
	})
}

func testAccApplicationDataSourceConfig_byInstance() string {
	return `
variable "identity_center_instance_arn" {
  type     = string
  nullable = false
}

resource "aws_accountaccess_application" "test" {
  identity_center_instance_arn = var.identity_center_instance_arn
}

data "aws_accountaccess_application" "test" {
  identity_center_instance_arn = aws_accountaccess_application.test.identity_center_instance_arn
}
`
}
