// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/accountaccess/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

// isNotFoundError reports whether err indicates the resource does not exist.
//
// The Account Access API is inconsistent about not-found signaling: some
// operations return ResourceNotFoundException, but others (observed on
// GetEntitlement / DeleteEntitlement against the preview service) return a
// ValidationException whose message contains "not found". Treat both as
// not-found so Read can remove the resource from state and Delete can no-op.
func isNotFoundError(err error) bool {
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return true
	}
	var ve *awstypes.ValidationException
	if errors.As(err, &ve) {
		return strings.Contains(strings.ToLower(aws.ToString(ve.Message)), "not found")
	}
	return false
}
