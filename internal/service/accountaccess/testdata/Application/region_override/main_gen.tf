# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_accountaccess_application" "test" {
  identity_center_instance_arn = var.identity_center_instance_arn
  region = var.region

}

variable "identity_center_instance_arn" {
  type     = string
  nullable = false
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
