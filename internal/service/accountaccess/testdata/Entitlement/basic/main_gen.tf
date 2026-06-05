# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_accountaccess_entitlement" "test" {
  application_arn = var.application_arn
  principal_id    = var.principal_id
  principal_type  = var.principal_type
  role_arn        = var.role_arn}

variable "application_arn" {
  type     = string
  nullable = false
}

variable "principal_id" {
  type     = string
  nullable = false
}

variable "principal_type" {
  type     = string
  default  = "USER"
  nullable = false
}

variable "role_arn" {
  type     = string
  nullable = false
}

variable "rName" {
  type     = string
  nullable = false
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
