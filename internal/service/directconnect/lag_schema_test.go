// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceLagSchemaOperationalFields(t *testing.T) {
	t.Parallel()

	resource := resourceLag()
	resource.Schema = resource.SchemaFunc()
	resource.SchemaFunc = nil

	if err := resource.InternalValidate(nil, true); err != nil {
		t.Fatalf("validating resource schema: %s", err)
	}

	testCases := map[string]schema.ValueType{
		"allows_hosted_connections": schema.TypeBool,
		"aws_device":                schema.TypeString,
		"aws_logical_device_id":     schema.TypeString,
		"lag_state":                 schema.TypeString,
	}

	for name, expectedType := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			attribute, ok := resource.Schema[name]
			if !ok {
				t.Fatalf("schema missing %q", name)
			}

			if attribute.Type != expectedType {
				t.Errorf("Type = %v, want %v", attribute.Type, expectedType)
			}

			if !attribute.Computed {
				t.Error("Computed = false, want true")
			}

			if attribute.Optional {
				t.Error("Optional = true, want false")
			}

			if attribute.Required {
				t.Error("Required = true, want false")
			}
		})
	}
}
