// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestResourceConnection_NameNotForceNew(t *testing.T) {
	t.Parallel()

	if resourceConnection().SchemaMap()[names.AttrName].ForceNew {
		t.Errorf("name schema ForceNew = true, want false")
	}
}
