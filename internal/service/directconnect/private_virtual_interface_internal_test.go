// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect

import "testing"

func TestValidatePrivateVirtualInterfaceConnectionMTUChange(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		isNew               bool
		connectionIDChanged bool
		mtuChanged          bool
		wantErr             bool
	}{
		"create with connection ID and MTU": {
			isNew:               true,
			connectionIDChanged: true,
			mtuChanged:          true,
		},
		"existing connection ID only": {
			connectionIDChanged: true,
		},
		"existing MTU only": {
			mtuChanged: true,
		},
		"existing connection ID and MTU": {
			connectionIDChanged: true,
			mtuChanged:          true,
			wantErr:             true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := validatePrivateVirtualInterfaceConnectionMTUChange(testCase.isNew, testCase.connectionIDChanged, testCase.mtuChanged)
			if gotErr := err != nil; gotErr != testCase.wantErr {
				t.Fatalf("error = %v, want error = %t", err, testCase.wantErr)
			}
		})
	}
}
