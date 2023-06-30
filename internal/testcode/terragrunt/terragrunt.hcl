terraform {
  source = "../modules"
}

# generate provider for aws
generate "provider" {
  path      = "provider.tf"
  if_exists = "overwrite_terragrunt"
  contents  = <<EOF
provider "aws" {
  region  = "eu-west-3"
}
EOF
}

inputs = {
  name = "zebi"
}
