// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// testAccPreCheck verifies that the test environment can reach the Account
// Access service. It calls ListApplications to confirm credentials are valid
// and the calling account is allowlisted for the preview.
func testAccPreCheck(ctx context.Context, t *testing.T) {
	acctest.PreCheckPartitionHasService(t, "account-access")

	conn := acctest.Provider.Meta().(*conns.AWSClient).AccountAccessClient(ctx)

	input := &accountaccess.ListApplicationsInput{}
	_, err := conn.ListApplications(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
