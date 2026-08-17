---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_virtual_interface_routes"
description: |-
  Provides details about routes for an AWS Direct Connect virtual interface.
---

# Data Source: aws_dx_virtual_interface_routes

Provides details about routes for an AWS Direct Connect virtual interface.

## Example Usage

### Basic Usage

```hcl
data "aws_dx_virtual_interface_routes" "example" {
  virtual_interface_id = "dxvif-abcde123"
}
```

## Argument Reference

The following arguments are required:

* `virtual_interface_id` - (Required) ID of the Direct Connect virtual interface.

The following arguments are optional:

* `address_family` - (Optional) Address family of routes to return. Allowed values are: `ipv4` and `ipv6`.
* `as_path` - (Optional) Autonomous system numbers used to filter routes by AS path.
* `cidrs` - (Optional) CIDR prefixes used to filter routes. A maximum of 10 CIDRs can be specified.
* `communities` - (Optional) BGP communities used to filter routes.
* `region` - (Optional) Region where this data source will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `route_direction` - (Optional) Direction of routes to return. Allowed values are: `accepted` and `advertised`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `routes` - Routes for the virtual interface. See [`routes`](#routes) below.

### `routes` Block

The `routes` block supports:

* `address_family` - Address family of the route.
* `as_path` - Autonomous system path of the route. See [`as_path`](#as_path) below.
* `aws_logical_device_id` - Direct Connect endpoint that terminates the logical connection.
* `cidr` - CIDR prefix of the route.
* `communities` - BGP communities associated with the route.
* `route_direction` - Direction of the route.
* `route_installed_at` - Time when the route was installed.

### `as_path` Block

The `as_path` block supports:

* `path` - Autonomous system numbers in the path segment.
* `path_type` - Type of the path segment.
