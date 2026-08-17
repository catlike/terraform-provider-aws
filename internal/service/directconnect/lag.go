// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package directconnect

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dx_lag", name="LAG")
// @Tags(identifierAttribute="arn")
func resourceLag() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLagCreate,
		ReadWithoutTimeout:   resourceLagRead,
		UpdateWithoutTimeout: resourceLagUpdate,
		DeleteWithoutTimeout: resourceLagDelete,

		CustomizeDiff: resourceLagCustomizeDiff,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrConnectionID: {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
				},
				"child_connection_tags": tftags.TagsSchemaForceNew(),
				"connection_ids": {
					Type:     schema.TypeSet,
					Computed: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"connections_bandwidth": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validConnectionBandWidth(),
				},
				names.AttrForceDestroy: {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"has_logical_redundancy": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"jumbo_frame_capable": {
					Type:     schema.TypeBool,
					Computed: true,
				},
				"number_of_connections": {
					Type:         schema.TypeInt,
					Optional:     true,
					Default:      0,
					ForceNew:     true,
					ValidateFunc: validation.IntBetween(0, 4),
				},
				names.AttrLocation: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
				},
				names.AttrOwnerAccountID: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrProviderName: {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
			}
		},
	}
}

func resourceLagCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, meta any) error {
	for _, key := range []string{"number_of_connections", "connections_bandwidth", names.AttrConnectionID, "child_connection_tags"} {
		if !diff.NewValueKnown(key) {
			return nil
		}
	}

	return validateLagProvisioningControls(
		diff.Get("number_of_connections").(int),
		diff.Get("connections_bandwidth").(string),
		diff.Get(names.AttrConnectionID).(string),
		diff.Get("child_connection_tags").(map[string]any),
	)
}

func validateLagProvisioningControls(numberOfConnections int, connectionsBandwidth, connectionID string, childConnectionTags map[string]any) error {
	if numberOfConnections > 0 && connectionID != "" {
		return fmt.Errorf("'number_of_connections' cannot be set with 'connection_id'")
	}

	if numberOfConnections > 0 && (connectionsBandwidth == "100Gbps" || connectionsBandwidth == "400Gbps") && numberOfConnections > 2 {
		return fmt.Errorf("'number_of_connections' cannot be greater than 2 when 'connections_bandwidth' is %q", connectionsBandwidth)
	}

	if numberOfConnections == 0 && len(childConnectionTags) > 0 {
		return fmt.Errorf("'child_connection_tags' can only be set when 'number_of_connections' is greater than 0")
	}

	return nil
}

func lagConnectionIDs(connections []awstypes.Connection) []string {
	connectionIDs := make([]string, 0, len(connections))

	for _, connection := range connections {
		if connectionID := aws.ToString(connection.ConnectionId); connectionID != "" {
			connectionIDs = append(connectionIDs, connectionID)
		}
	}

	return connectionIDs
}

func resourceLagCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	name := d.Get(names.AttrName).(string)
	numberOfConnections := d.Get("number_of_connections").(int)
	connectionID, connectionIDSpecified := d.GetOk(names.AttrConnectionID)
	connectionIDString := ""
	if connectionIDSpecified {
		connectionIDString = connectionID.(string)
	}
	childConnectionTags := d.Get("child_connection_tags").(map[string]any)

	if err := validateLagProvisioningControls(numberOfConnections, d.Get("connections_bandwidth").(string), connectionIDString, childConnectionTags); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &directconnect.CreateLagInput{
		ConnectionsBandwidth: aws.String(d.Get("connections_bandwidth").(string)),
		LagName:              aws.String(name),
		Location:             aws.String(d.Get(names.AttrLocation).(string)),
		Tags:                 getTagsIn(ctx),
	}

	if numberOfConnections > 0 {
		input.NumberOfConnections = int32(numberOfConnections)
		input.ChildConnectionTags = svcTags(tftags.New(ctx, childConnectionTags))
	} else {
		// Direct Connect requires one connection, while zero preserves the historical
		// provider behavior of creating and then deleting a bootstrap connection.
		input.NumberOfConnections = 1
	}

	if connectionIDSpecified {
		input.ConnectionId = aws.String(connectionIDString)
	}

	if v, ok := d.GetOk(names.AttrProviderName); ok {
		input.ProviderName = aws.String(v.(string))
	}

	output, err := conn.CreateLag(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Direct Connect LAG (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.LagId))

	if numberOfConnections == 0 && !connectionIDSpecified {
		// Delete the compatibility bootstrap connections so legacy configurations
		// continue to create an empty LAG.
		for _, connection := range output.Connections {
			if err := deleteConnection(ctx, conn, aws.ToString(connection.ConnectionId), waitConnectionDeleted); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
		if err := d.Set("connection_ids", []string{}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting Direct Connect LAG (%s) connection IDs: %s", d.Id(), err)
		}
	} else if numberOfConnections > 0 {
		lag, err := waitLagConnectionsVisible(ctx, conn, d.Id(), numberOfConnections)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect LAG (%s) connections: %s", d.Id(), err)
		}
		if err := d.Set("connection_ids", lagConnectionIDs(lag.Connections)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting Direct Connect LAG (%s) connection IDs: %s", d.Id(), err)
		}
	} else {
		if err := d.Set("connection_ids", []string{}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting Direct Connect LAG (%s) connection IDs: %s", d.Id(), err)
		}
	}

	return append(diags, resourceLagRead(ctx, d, meta)...)
}

func resourceLagRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	lag, err := findLagByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Direct Connect LAG (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect LAG (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Region:    aws.ToString(lag.Region),
		Service:   "directconnect",
		AccountID: aws.ToString(lag.OwnerAccount),
		Resource:  fmt.Sprintf("dxlag/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("connections_bandwidth", lag.ConnectionsBandwidth)
	d.Set("has_logical_redundancy", lag.HasLogicalRedundancy)
	d.Set("jumbo_frame_capable", lag.JumboFrameCapable)
	d.Set(names.AttrLocation, lag.Location)
	d.Set(names.AttrName, lag.LagName)
	d.Set(names.AttrOwnerAccountID, lag.OwnerAccount)
	d.Set(names.AttrProviderName, lag.ProviderName)

	trackedConnectionIDs := d.Get("connection_ids").(*schema.Set)
	lagConnectionIDs := make(map[string]struct{}, len(lag.Connections))
	for _, connection := range lag.Connections {
		lagConnectionIDs[aws.ToString(connection.ConnectionId)] = struct{}{}
	}

	connectionIDs := make([]string, 0, trackedConnectionIDs.Len())
	for _, connectionID := range trackedConnectionIDs.List() {
		connectionID := connectionID.(string)
		if connectionID == "" {
			continue
		}

		if _, ok := lagConnectionIDs[connectionID]; ok {
			connectionIDs = append(connectionIDs, connectionID)
			continue
		}

		_, err := findConnectionByID(ctx, conn, connectionID)
		if retry.NotFound(err) {
			continue
		}
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Direct Connect Connection (%s): %s", connectionID, err)
		}

		connectionIDs = append(connectionIDs, connectionID)
	}
	if err := d.Set("connection_ids", connectionIDs); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Direct Connect LAG (%s) connection IDs: %s", d.Id(), err)
	}

	return diags
}

func resourceLagUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	if d.HasChange(names.AttrName) {
		input := &directconnect.UpdateLagInput{
			LagId:   aws.String(d.Id()),
			LagName: aws.String(d.Get(names.AttrName).(string)),
		}

		_, err := conn.UpdateLag(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Direct Connect LAG (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceLagRead(ctx, d, meta)...)
}

func resourceLagDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	if d.Get(names.AttrForceDestroy).(bool) {
		lag, err := findLagByID(ctx, conn, d.Id())

		if retry.NotFound(err) {
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Direct Connect LAG (%s): %s", d.Id(), err)
		}

		for _, connection := range lag.Connections {
			if err := deleteConnection(ctx, conn, aws.ToString(connection.ConnectionId), waitConnectionDeleted); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	} else {
		if v, ok := d.GetOk("connection_ids"); ok {
			for _, connectionID := range v.(*schema.Set).List() {
				if err := deleteConnection(ctx, conn, connectionID.(string), waitConnectionDeleted); err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
			}
		}

		if v, ok := d.GetOk(names.AttrConnectionID); ok {
			if err := deleteConnectionLAGAssociation(ctx, conn, v.(string), d.Id()); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	log.Printf("[DEBUG] Deleting Direct Connect LAG: %s", d.Id())
	input := directconnect.DeleteLagInput{
		LagId: aws.String(d.Id()),
	}
	_, err := conn.DeleteLag(ctx, &input)

	if errs.IsAErrorMessageContains[*awstypes.DirectConnectClientException](err, "Could not find Lag with ID") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Direct Connect LAG (%s): %s", d.Id(), err)
	}

	if _, err := waitLagDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect LAG (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findLagByID(ctx context.Context, conn *directconnect.Client, id string) (*awstypes.Lag, error) {
	input := &directconnect.DescribeLagsInput{
		LagId: aws.String(id),
	}
	output, err := findLag(ctx, conn, input, tfslices.PredicateTrue[*awstypes.Lag]())

	if err != nil {
		return nil, err
	}

	if state := output.LagState; state == awstypes.LagStateDeleted {
		return nil, &retry.NotFoundError{
			Message: string(state),
		}
	}

	return output, nil
}

func findLag(ctx context.Context, conn *directconnect.Client, input *directconnect.DescribeLagsInput, filter tfslices.Predicate[*awstypes.Lag]) (*awstypes.Lag, error) {
	output, err := findLags(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findLags(ctx context.Context, conn *directconnect.Client, input *directconnect.DescribeLagsInput, filter tfslices.Predicate[*awstypes.Lag]) ([]awstypes.Lag, error) {
	output, err := conn.DescribeLags(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.DirectConnectClientException](err, "Could not find Lag with ID") {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return tfslices.Filter(output.Lags, tfslices.PredicateValue(filter)), nil
}

func statusLagConnectionsVisible(conn *directconnect.Client, id string, expectedCount int) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		lag, err := findLagByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "pending", nil
		}
		if err != nil {
			return nil, "", err
		}

		if len(lag.Connections) == expectedCount {
			return lag, "visible", nil
		}

		return lag, "pending", nil
	}
}

func waitLagConnectionsVisible(ctx context.Context, conn *directconnect.Client, id string, expectedCount int) (*awstypes.Lag, error) {
	const timeout = 2 * time.Minute

	stateConf := &retry.StateChangeConf{
		Pending:                   []string{"pending"},
		Target:                    []string{"visible"},
		Refresh:                   statusLagConnectionsVisible(conn, id, expectedCount),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Lag); ok {
		return output, err
	}

	return nil, err
}

func statusLag(conn *directconnect.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findLagByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.LagState), nil
	}
}

func waitLagDeleted(ctx context.Context, conn *directconnect.Client, id string) (*awstypes.Lag, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LagStateAvailable, awstypes.LagStateRequested, awstypes.LagStatePending, awstypes.LagStateDeleting),
		Target:  []string{},
		Refresh: statusLag(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Lag); ok {
		return output, err
	}

	return nil, err
}
