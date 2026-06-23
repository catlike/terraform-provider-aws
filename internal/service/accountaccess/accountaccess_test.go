// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// Account Access launches in us-east-1 and us-west-2 only during preview.
// Tests gate on these via testAccPreCheckRegion.
var accountAccessRegions = []string{
	endpoints.UsEast1RegionID,
	endpoints.UsWest2RegionID,
}

// testAccPreCheck verifies the test environment can reach Account Access and
// that the prerequisite IAM Identity Center instance exists in the test
// account. It is the standard PreCheck for every acceptance test in this
// package.
//
// Prerequisites that CANNOT be provisioned by Terraform in-test (and so are
// asserted here instead):
//   - The calling account must be allowlisted for the Account Access preview.
//   - IAM Identity Center must be enabled with an instance in the test region
//     (org-level setup). PreCheckSSOAdminInstances asserts this.
func testAccPreCheck(ctx context.Context, t *testing.T) {
	acctest.PreCheckSSOAdminInstances(ctx, t)

	conn := acctest.Provider.Meta().(*conns.AWSClient).AccountAccessClient(ctx)

	_, err := conn.ListApplications(ctx, &accountaccess.ListApplicationsInput{})
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

// testAccPreCheckRegion skips the test unless the configured region is one
// where the Account Access preview is available.
func testAccPreCheckRegion(t *testing.T) {
	acctest.PreCheckRegion(t, accountAccessRegions...)
}

// testAccPrerequisitesConfig returns HCL that self-provisions everything an
// Account Access acceptance test needs, EXCEPT the IAM Identity Center
// instance (which is an org-level prerequisite discovered via data source):
//
//   - data.aws_ssoadmin_instances — the IdC instance ARN + identity store ID
//   - aws_identitystore_user       — a USER principal
//   - aws_identitystore_group      — a GROUP principal
//   - aws_iam_role                 — a target role with the Account Access
//     trust policy (account-access-preview.amazonaws.com + sts:AssumeRole,
//     sts:SetContext, sts:TagSession — see CONTEXT.md §4/§5)
//
// Outputs are referenced by callers as:
//
//	data.aws_ssoadmin_instances.test
//	aws_identitystore_user.test.user_id
//	aws_identitystore_group.test.group_id
//	aws_iam_role.test.arn
//
// NOTE: the trust-policy service principal is the PREVIEW principal
// (account-access-preview.amazonaws.com). At GA this likely becomes
// account-access.amazonaws.com — update here when confirmed.
func testAccPrerequisitesConfig(rName string) string {
	return acctest.ConfigCompose(`
data "aws_ssoadmin_instances" "test" {}

locals {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  instance_arn      = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_identitystore_user" "test" {
  identity_store_id = local.identity_store_id

  display_name = "` + rName + `"
  user_name    = "` + rName + `"

  name {
    given_name  = "Acceptance"
    family_name = "Test"
  }

  emails {
    value = "` + rName + `@example.com"
  }
}

resource "aws_identitystore_group" "test" {
  identity_store_id = local.identity_store_id
  display_name      = "` + rName + `"
  description       = "Account Access acceptance test group"
}

resource "aws_iam_role" "test" {
  name = "` + rName + `"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "account-access-preview.amazonaws.com"
        }
        Action = [
          "sts:AssumeRole",
          "sts:SetContext",
          "sts:TagSession",
        ]
      },
    ]
  })
}
`)
}
