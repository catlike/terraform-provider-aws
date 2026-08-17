---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_lag"
description: |-
  Provides a Direct Connect LAG.
---

# Resource: aws_dx_lag

Provides a Direct Connect LAG. Connections can be added to the LAG via the [`aws_dx_connection`](/docs/providers/aws/r/dx_connection.html) and [`aws_dx_connection_association`](/docs/providers/aws/r/dx_connection_association.html) resources.

~> *NOTE:* `number_of_connections` defaults to `0` for backward compatibility. With this default, Direct Connect creates one bootstrap connection and Terraform immediately deletes it, leaving an empty LAG as in earlier provider versions. Set a positive value to retain that many automatically created child connections.

~> *WARNING:* Terraform manages only the automatically created child connections recorded in `connection_ids`. It does not adopt other connections that are later associated with the LAG. With `force_destroy = false`, destroying this resource deletes the tracked child connections but leaves externally associated connections in place, so Direct Connect can reject LAG deletion until they are removed or disassociated. With `force_destroy = true`, Terraform retains the existing behavior of deleting all current LAG connections.

## Example Usage

```terraform
resource "aws_dx_lag" "hoge" {
  name                  = "tf-dx-lag"
  connections_bandwidth = "1Gbps"
  location              = "EqDC2"
  force_destroy         = true
}
```

### Retain automatically created connections

```terraform
resource "aws_dx_lag" "provisioned" {
  name                  = "tf-dx-lag-provisioned"
  connections_bandwidth = "10Gbps"
  location              = "EqDC2"
  number_of_connections = 2

  child_connection_tags = {
    Environment = "test"
    ManagedBy   = "Terraform"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) The name of the LAG.
* `connections_bandwidth` - (Required) The bandwidth of the individual dedicated connections bundled by the LAG. Valid values: 1Gbps, 10Gbps, 100Gbps, and 400Gbps. Case sensitive. Refer to the AWS Direct Connection supported bandwidths for [Dedicated Connections](https://docs.aws.amazon.com/directconnect/latest/UserGuide/dedicated_connection.html).
* `location` - (Required) The AWS Direct Connect location in which the LAG should be allocated. See [DescribeLocations](https://docs.aws.amazon.com/directconnect/latest/APIReference/API_DescribeLocations.html) for the list of AWS Direct Connect locations. Use `locationCode`.
* `connection_id` - (Optional) The ID of an existing dedicated connection to migrate to the LAG. Cannot be specified with a positive `number_of_connections`.
* `number_of_connections` - (Optional, Forces new resource, Default: `0`) The number of dedicated connections Direct Connect creates and bundles with the LAG. Valid values are `0` through `4`. `0` preserves the historical empty-LAG behavior; a positive value retains exactly that many automatically created child connections. For `100Gbps` and `400Gbps`, the maximum is `2`.
* `child_connection_tags` - (Optional, Forces new resource) A map of tags to apply to automatically created child connections. This argument requires a positive `number_of_connections` and does not inherit provider `default_tags`.
* `force_destroy` - (Optional, Default:false) A boolean that indicates all connections associated with the LAG should be deleted so that the LAG can be destroyed without error. These objects are *not* recoverable. When false, Terraform deletes only child connections listed in `connection_ids` before attempting LAG deletion.
* `provider_name` - (Optional) The name of the service provider associated with the LAG.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the LAG.
* `connection_ids` - The IDs of automatically created child connections retained and managed by this resource. Connections associated outside this resource are never adopted.
* `has_logical_redundancy` - Indicates whether the LAG supports a secondary BGP peer in the same address family (IPv4/IPv6).
* `id` - The ID of the LAG.
* `jumbo_frame_capable` -Indicates whether jumbo frames (9001 MTU) are supported.
* `owner_account_id` - The ID of the AWS account that owns the LAG.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Direct Connect LAGs using the LAG `id`. For example:

```terraform
import {
  to = aws_dx_lag.test_lag
  id = "dxlag-fgnsp5rq"
}
```

Imported LAGs use the compatibility defaults: `number_of_connections = 0`, empty `child_connection_tags`, and empty `connection_ids`. Terraform cannot determine which existing LAG connections were automatically created by a prior resource instance, so it does not adopt or delete them unless `force_destroy = true`.

Using `terraform import`, import Direct Connect LAGs using the LAG `id`. For example:

```console
% terraform import aws_dx_lag.test_lag dxlag-fgnsp5rq
```
