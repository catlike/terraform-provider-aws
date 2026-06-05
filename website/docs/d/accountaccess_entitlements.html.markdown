---
subcategory: "Account Access"
layout: "aws"
page_title: "AWS: aws_accountaccess_entitlements"
description: |-
  Lists AWS Account Access Entitlements matching a required filter.
---

# Data Source: aws_accountaccess_entitlements

Lists AWS Account Access [Entitlements](../../resources/accountaccess_entitlement) for a given Application, filtered by principal, role, or target account.

~> **Note:** The Account Access `ListEntitlements` API requires both the Application ARN and at least one filter dimension — there is no list-all. You must supply at least one of `principal_id`+`principal_type`, `role_arn`, or `account_id`.

## Example Usage

### Filter by Principal

```terraform
data "aws_accountaccess_entitlements" "user" {
  application_arn = aws_accountaccess_application.example.arn
  principal_id    = "11111111-2222-3333-4444-555555555555"
  principal_type  = "USER"
}
```

### Filter by Role

```terraform
data "aws_accountaccess_entitlements" "developer_role" {
  application_arn = aws_accountaccess_application.example.arn
  role_arn        = "arn:aws:iam::123456789012:role/Developer"
}
```

### Filter by Target Account

```terraform
data "aws_accountaccess_entitlements" "target" {
  application_arn = aws_accountaccess_application.example.arn
  account_id      = "123456789012"
}
```

## Argument Reference

The following arguments are required:

* `application_arn` - (Required) ARN of the parent Application to list entitlements within.

You must also supply **at least one** of the following filter arguments:

* `principal_id` - (Optional) Identity Center user or group ID to filter by. Must be set together with `principal_type`.
* `principal_type` - (Optional) Type of principal. Valid values: `USER`, `GROUP`. Must be set together with `principal_id`.
* `role_arn` - (Optional) Target IAM role ARN to filter by.
* `account_id` - (Optional) 12-digit AWS account ID to filter by.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `entitlements` - List of matching Entitlements. Each element contains:
    * `account_id` - 12-digit AWS account ID extracted from the target role ARN.
    * `account_name` - Best-effort human-readable name of the target account.
    * `created_at` - Date and time, in [RFC3339 format](https://datatracker.ietf.org/doc/html/rfc3339), when the Entitlement was created.
    * `entitlement_id` - Service-assigned unique identifier for the Entitlement.
    * `principal_id` - Identity Center user or group ID granted access.
    * `principal_type` - Type of principal: `USER` or `GROUP`.
    * `role_arn` - Target IAM role ARN.
* `id` - Application ARN (synthetic — kept for compatibility with Terraform's data-source ID convention).
