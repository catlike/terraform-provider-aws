---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_virtual_interfaces"
description: Provides details about Direct Connect virtual interfaces in the current AWS Region.
---

# Data Source: aws_dx_virtual_interfaces

Provides details about every private, public, transit, and hosted Direct Connect virtual interface visible in the configured AWS Region.

This data source intentionally exposes no filter arguments. Use a Terraform `for` expression to select the virtual interfaces needed by your configuration.

## Example Usage

```hcl
data "aws_dx_virtual_interfaces" "example" {}

locals {
  private_virtual_interfaces = [
    for vif in data.aws_dx_virtual_interfaces.example.virtual_interfaces : vif
    if vif.type == "private" && vif.state == "available"
  ]
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `virtual_interfaces` - List of virtual interface inventory records. Each object contains:
  * `address_family` - Address family for the BGP peer.
  * `amazon_address` - IP address assigned to the Amazon interface.
  * `amazon_side_asn` - Autonomous system number for the Amazon side of the connection.
  * `arn` - ARN of the virtual interface.
  * `aws_device` - Direct Connect endpoint that terminates the physical connection.
  * `aws_logical_device_id` - Direct Connect endpoint that terminates the logical connection.
  * `bgp_asn` - BGP autonomous system number.
  * `bgp_asn_long` - Long BGP autonomous system number.
  * `connection_id` - ID of the Direct Connect connection.
  * `customer_address` - IP address assigned to the customer interface.
  * `direct_connect_gateway_id` - ID of the Direct Connect gateway.
  * `id` - ID of the virtual interface.
  * `jumbo_frame_capable` - Whether jumbo frames are supported.
  * `location` - Location of the connection.
  * `mtu` - Maximum transmission unit, in bytes.
  * `name` - Name of the virtual interface.
  * `owner_account_id` - AWS account ID that owns the virtual interface.
  * `rate_limit` - Rate limit associated with the virtual interface, when available.
  * `route_filter_prefixes` - CIDR prefixes advertised to AWS. Applies to public virtual interfaces.
  * `site_link_enabled` - Whether SiteLink is enabled.
  * `state` - State of the virtual interface.
  * `tags` - Map of tags associated with the virtual interface, after provider `ignore_tags` configuration is applied.
  * `type` - Type of virtual interface: `private`, `public`, or `transit`.
  * `virtual_gateway_id` - ID of the virtual private gateway.
  * `vlan` - VLAN ID.

~> **Note:** This inventory data source intentionally does not expose authentication keys, BGP peer authentication material, or customer router configuration. Use the dedicated [`aws_dx_router_configuration`](/docs/providers/aws/d/dx_router_configuration.html) data source only when router configuration is explicitly required.
