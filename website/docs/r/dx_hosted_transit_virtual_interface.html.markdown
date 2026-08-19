---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_hosted_transit_virtual_interface"
description: |-
  Provides a Direct Connect hosted transit virtual interface resource.
---

# Resource: aws_dx_hosted_transit_virtual_interface

Provides a Direct Connect hosted transit virtual interface resource.
This resource represents the allocator's side of the hosted virtual interface.
A hosted virtual interface is a virtual interface that is owned by another AWS account.

## Example Usage

```terraform
resource "aws_dx_hosted_transit_virtual_interface" "example" {
  connection_id = aws_dx_connection.example.id

  name           = "tf-transit-vif-example"
  vlan           = 4094
  address_family = "ipv4"
  bgp_asn        = 65352
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `address_family` - (Required) The address family for the BGP peer. `ipv4 ` or `ipv6`.
* `bgp_asn` - (Required) The autonomous system (AS) number for Border Gateway Protocol (BGP) configuration.
* `connection_id` - (Required) ID of the Direct Connect connection or LAG hosting the virtual interface. Updating this argument reassociates the existing hosted virtual interface without changing its ID.
* `name` - (Required) Name for the virtual interface. Updating this argument performs an in-place update after the owner account accepts the hosted virtual interface.
* `owner_account_id` - (Required) The AWS account that will own the new virtual interface.
* `vlan` - (Required) The VLAN ID.
* `amazon_address` - (Optional) The IPv4 CIDR address to use to send traffic to Amazon. Required for IPv4 BGP peers.
* `bgp_auth_key` - (Optional) The authentication key for BGP configuration.
* `customer_address` - (Optional) The IPv4 CIDR destination address to which Amazon should send traffic. Required for IPv4 BGP peers.
* `mtu` - (Optional) The maximum transmission unit (MTU) is the size, in bytes, of the largest permissible packet that can be passed over the connection. The MTU of a virtual transit interface can be either `1500` or `8500` (jumbo frames). Default is `1500`.

### Reassociation

Updating `connection_id` reassociates the existing hosted transit virtual interface; it does not create a new VIF. The allocator account managing this resource must own the target connection or LAG and either the VIF or its current connection. `owner_account_id` remains the account that owns the hosted VIF and cannot be changed. The owner account must accept the allocation (for example, with [`aws_dx_hosted_transit_virtual_interface_accepter`](dx_hosted_transit_virtual_interface_accepter.html.markdown)); its Direct Connect gateway association remains in place while the allocator moves the VIF.

Before changing `connection_id`, ensure that the hosted VIF and target are available for reassociation. The target must be compatible with the VIF's existing VLAN and BGP peer IP addressing; a conflicting VLAN or Amazon/customer IP address prevents the reassociation. This resource does not change those allocation settings during reassociation.

Reassociation can interrupt connectivity while the VIF moves and BGP sessions reconverge. Plan for traffic loss and verify the BGP configuration and reachability on the target before applying the change.

A LAG ID can be used as the target only when the VIF is not associated with a hosted connection. A VIF on a hosted connection cannot be moved to a LAG; migrate the hosted connection and its VIFs with `AssociateHostedConnection` instead.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the virtual interface.
* `arn` - The ARN of the virtual interface.
* `aws_device` - The Direct Connect endpoint on which the virtual interface terminates.
* `jumbo_frame_capable` - Indicates whether jumbo frames (8500 MTU) are supported.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `update` - (Default `30m`)
- `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Direct Connect hosted transit virtual interfaces using the VIF `id`. For example:

```terraform
import {
  to = aws_dx_hosted_transit_virtual_interface.test
  id = "dxvif-33cc44dd"
}
```

Using `terraform import`, import Direct Connect hosted transit virtual interfaces using the VIF `id`. For example:

```console
% terraform import aws_dx_hosted_transit_virtual_interface.test dxvif-33cc44dd
```
