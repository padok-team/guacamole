# TG_INP_001 OK: a registry source can't be read locally, the layer is skipped.
terraform {
  source = "terraform-aws-modules/vpc/aws"
}

inputs = {
  name = "my-vpc"
}
