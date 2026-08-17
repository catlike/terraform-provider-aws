// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestValidateLagProvisioningControls(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		numberOfConnections  int
		connectionsBandwidth string
		connectionID         string
		childConnectionTags  map[string]any
		wantErr              bool
	}{
		"compatibility sentinel": {
			numberOfConnections:  0,
			connectionsBandwidth: "10Gbps",
		},
		"retained children": {
			numberOfConnections:  2,
			connectionsBandwidth: "10Gbps",
			childConnectionTags:  map[string]any{"Environment": "test"},
		},
		"existing connection": {
			numberOfConnections:  0,
			connectionsBandwidth: "10Gbps",
			connectionID:         "dxcon-12345678",
		},
		"connection ID with retained children": {
			numberOfConnections:  1,
			connectionsBandwidth: "10Gbps",
			connectionID:         "dxcon-12345678",
			wantErr:              true,
		},
		"too many 100Gbps children": {
			numberOfConnections:  3,
			connectionsBandwidth: "100Gbps",
			wantErr:              true,
		},
		"too many 400Gbps children": {
			numberOfConnections:  3,
			connectionsBandwidth: "400Gbps",
			wantErr:              true,
		},
		"child tags without retained children": {
			numberOfConnections:  0,
			connectionsBandwidth: "10Gbps",
			childConnectionTags:  map[string]any{"Environment": "test"},
			wantErr:              true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := validateLagProvisioningControls(testCase.numberOfConnections, testCase.connectionsBandwidth, testCase.connectionID, testCase.childConnectionTags)
			if gotErr := err != nil; gotErr != testCase.wantErr {
				t.Fatalf("validateLagProvisioningControls() error = %v, wantErr %t", err, testCase.wantErr)
			}
		})
	}
}

func TestResourceLagSchema(t *testing.T) {
	t.Parallel()

	resource := resourceLag()
	resourceSchema := resource.SchemaFunc()
	if err := resource.InternalValidate(resourceSchema, true); err != nil {
		t.Fatalf("resource.InternalValidate(): %v", err)
	}

	for _, name := range []string{"number_of_connections", "child_connection_tags"} {
		attribute, ok := resourceSchema[name]
		if !ok {
			t.Fatalf("schema missing %q", name)
		}
		if !attribute.ForceNew {
			t.Errorf("schema %q ForceNew = false, want true", name)
		}
	}

	numberOfConnections := resourceSchema["number_of_connections"]
	if !numberOfConnections.Optional || numberOfConnections.Default != 0 {
		t.Errorf("number_of_connections Optional/Default = %t/%#v, want true/0", numberOfConnections.Optional, numberOfConnections.Default)
	}
	if warnings, errors := numberOfConnections.ValidateFunc(5, "number_of_connections"); len(warnings) != 0 || len(errors) == 0 {
		t.Errorf("number_of_connections validator warnings/errors = %v/%v, want no warnings and an error", warnings, errors)
	}

	childConnectionTags := resourceSchema["child_connection_tags"]
	if !childConnectionTags.Optional || childConnectionTags.Type != schema.TypeMap {
		t.Errorf("child_connection_tags Optional/Type = %t/%v, want true/schema.TypeMap", childConnectionTags.Optional, childConnectionTags.Type)
	}

	connectionIDs := resourceSchema["connection_ids"]
	if !connectionIDs.Computed || connectionIDs.Type != schema.TypeSet {
		t.Errorf("connection_ids Computed/Type = %t/%v, want true/schema.TypeSet", connectionIDs.Computed, connectionIDs.Type)
	}
}
