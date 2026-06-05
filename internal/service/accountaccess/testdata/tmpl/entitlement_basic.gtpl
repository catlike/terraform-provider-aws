resource "aws_accountaccess_entitlement" "test" {
  application_arn = var.application_arn
  principal_id    = var.principal_id
  principal_type  = var.principal_type
  role_arn        = var.role_arn

{{- template "region" -}}
}

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
