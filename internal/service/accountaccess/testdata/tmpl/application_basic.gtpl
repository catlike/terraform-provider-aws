resource "aws_accountaccess_application" "test" {
  identity_center_instance_arn = var.identity_center_instance_arn

{{- template "region" -}}
{{- template "tags" . }}
}

variable "identity_center_instance_arn" {
  type     = string
  nullable = false
}
