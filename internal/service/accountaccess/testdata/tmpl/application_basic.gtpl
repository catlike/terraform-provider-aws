data "aws_ssoadmin_instances" "test" {}

resource "aws_accountaccess_application" "test" {
  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
{{- template "region" -}}
{{- template "tags" . }}
}
