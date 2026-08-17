// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceLagMinimumLinksSchema(t *testing.T) {
	t.Parallel()

	minimumLinksSchema, ok := resourceLag().SchemaFunc()["minimum_links"]
	if !ok {
		t.Fatal("minimum_links schema is missing")
	}

	if minimumLinksSchema.Type != schema.TypeInt {
		t.Errorf("minimum_links type = %v, want %v", minimumLinksSchema.Type, schema.TypeInt)
	}

	if !minimumLinksSchema.Optional {
		t.Error("minimum_links is not optional")
	}

	if !minimumLinksSchema.Computed {
		t.Error("minimum_links is not computed")
	}

	for _, value := range []int{1, 4} {
		if _, errors := minimumLinksSchema.ValidateFunc(value, "minimum_links"); len(errors) != 0 {
			t.Errorf("minimum_links value %d returned validation errors: %v", value, errors)
		}
	}

	for _, value := range []int{0, 5} {
		if _, errors := minimumLinksSchema.ValidateFunc(value, "minimum_links"); len(errors) == 0 {
			t.Errorf("minimum_links value %d returned no validation errors", value)
		}
	}
}
