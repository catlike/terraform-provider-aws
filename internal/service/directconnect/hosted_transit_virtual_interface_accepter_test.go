// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceHostedTransitVirtualInterfaceAccepterSchema(t *testing.T) {
	t.Parallel()

	resource := resourceHostedTransitVirtualInterfaceAccepter()
	if err := resource.InternalValidate(nil, true); err != nil {
		t.Fatalf("invalid resource schema: %s", err)
	}

	attribute, ok := resource.SchemaMap()["sitelink_enabled"]
	if !ok {
		t.Fatal("expected sitelink_enabled schema")
	}
	if attribute.Type != schema.TypeBool {
		t.Errorf("expected sitelink_enabled to be TypeBool, got %s", attribute.Type)
	}
	if !attribute.Optional {
		t.Error("expected sitelink_enabled to be optional")
	}
	if !attribute.Computed {
		t.Error("expected sitelink_enabled to be computed")
	}
	if resource.Timeouts.Update == nil || *resource.Timeouts.Update != 10*time.Minute {
		t.Errorf("expected update timeout to be 10m, got %v", resource.Timeouts.Update)
	}
}
