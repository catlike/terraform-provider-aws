// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_network_interface_permission", name="Network Interface Permission")
func resourceNetworkInterfacePermission() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNetworkInterfacePermissionCreate,
		ReadWithoutTimeout:   resourceNetworkInterfacePermissionRead,
		DeleteWithoutTimeout: resourceNetworkInterfacePermissionDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"network_interface_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"aws_account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"permission": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.InterfacePermissionType](),
			},
			"permission_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceNetworkInterfacePermissionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	networkInterfaceID := d.Get("network_interface_id").(string)
	permission := types.InterfacePermissionType(d.Get("permission").(string))
	awsAccountID := d.Get("aws_account_id").(string)

	input := &ec2.CreateNetworkInterfacePermissionInput{
		NetworkInterfaceId: aws.String(networkInterfaceID),
		Permission:         permission,
		AwsAccountId:       aws.String(awsAccountID),
	}

	log.Printf("[DEBUG] Creating EC2 Network Interface Permission: %s", input)
	output, err := conn.CreateNetworkInterfacePermission(ctx, input)

	if err != nil {
		return diag.Errorf("creating EC2 Network Interface Permission: %s", err)
	}

	d.SetId(aws.ToString(output.NetworkInterfacePermission.NetworkInterfacePermissionId))

	if _, err := waitNetworkInterfacePermissionCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for EC2 Network Interface Permission (%s) create: %s", d.Id(), err)
	}

	return resourceNetworkInterfacePermissionRead(ctx, d, meta)
}

func resourceNetworkInterfacePermissionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	permission, err := findNetworkInterfacePermissionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Network Interface Permission %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading EC2 Network Interface Permission (%s): %s", d.Id(), err)
	}

	d.Set("permission_id", permission.NetworkInterfacePermissionId)
	d.Set("network_interface_id", permission.NetworkInterfaceId)
	d.Set("aws_account_id", permission.AwsAccountId)
	d.Set("permission", permission.Permission)

	return nil
}

func resourceNetworkInterfacePermissionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting EC2 Network Interface Permission: %s", d.Id())
	_, err := conn.DeleteNetworkInterfacePermission(ctx, &ec2.DeleteNetworkInterfacePermissionInput{
		NetworkInterfacePermissionId: aws.String(d.Id()),
	})

	if errs.IsA[*types.ClientError](err) && strings.Contains(err.Error(), "InvalidNetworkInterfacePermissionID.NotFound") {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting EC2 Network Interface Permission (%s): %s", d.Id(), err)
	}

	if _, err := waitNetworkInterfacePermissionDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for EC2 Network Interface Permission (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func findNetworkInterfacePermissionByID(ctx context.Context, conn *ec2.Client, id string) (*types.NetworkInterfacePermission, error) {
	input := &ec2.DescribeNetworkInterfacePermissionsInput{
		NetworkInterfacePermissionIds: []string{id},
	}

	output, err := conn.DescribeNetworkInterfacePermissions(ctx, input)

	if errs.IsA[*types.ClientError](err) && strings.Contains(err.Error(), "InvalidNetworkInterfacePermissionID.NotFound") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.NetworkInterfacePermissions) == 0 || output.NetworkInterfacePermissions[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.NetworkInterfacePermissions); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return &output.NetworkInterfacePermissions[0], nil
}

func statusNetworkInterfacePermission(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findNetworkInterfacePermissionByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.PermissionState), nil
	}
}

func waitNetworkInterfacePermissionCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.NetworkInterfacePermission, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(types.InterfacePermissionStatePending)},
		Target:  []string{string(types.InterfacePermissionStateGranted)},
		Refresh: statusNetworkInterfacePermission(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.NetworkInterfacePermission); ok {
		return output, err
	}

	return nil, err
}

func waitNetworkInterfacePermissionDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.NetworkInterfacePermission, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(types.InterfacePermissionStateGranted), string(types.InterfacePermissionStatePending)},
		Target:  []string{},
		Refresh: statusNetworkInterfacePermission(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.NetworkInterfacePermission); ok {
		return output, err
	}

	return nil, err
}
