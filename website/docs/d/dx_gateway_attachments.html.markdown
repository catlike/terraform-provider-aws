---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_gateway_attachments"
description: Provides details about attachments between Direct Connect gateways and virtual interfaces.
---

# Data Source: aws_dx_gateway_attachments

Provides details about attachments between Direct Connect gateways and virtual interfaces.

## Example Usage

```hcl
data "aws_dx_gateway_attachments" "example" {
  direct_connect_gateway_id = aws_dx_gateway.example.id
}
```

## Argument Reference

This data source supports the following arguments:

* `direct_connect_gateway_id` - (Optional) ID of the Direct Connect gateway. At least one of `direct_connect_gateway_id` or `virtual_interface_id` must be specified.
* `virtual_interface_id` - (Optional) ID of the virtual interface. At least one of `direct_connect_gateway_id` or `virtual_interface_id` must be specified.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `attachments` - List of matching Direct Connect gateway attachments, sorted by Direct Connect gateway ID and then virtual interface ID. Each attachment has the following attributes:
    * `attachment_state` - State of the attachment.
    * `attachment_type` - Type of the attachment.
    * `direct_connect_gateway_id` - ID of the Direct Connect gateway.
    * `state_change_error` - Error message when the attachment state could not advance.
    * `virtual_interface_id` - ID of the virtual interface.
    * `virtual_interface_owner_account_id` - AWS account ID that owns the virtual interface.
    * `virtual_interface_region` - AWS Region where the virtual interface is located.

~> **Note:** Detached attachment records can appear while a Direct Connect gateway or virtual interface is transitioning.
