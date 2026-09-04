terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# Data source depends only on a variable and static values, so it is read during
# the plan (TF_DAT_001 OK).
data "aws_ami" "this" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = [var.ami_name_pattern]
  }
}

# Unique resource type -> named "this" (TF_NAM_001), snake_case (TF_NAM_002),
# no stuttering (TF_NAM_003).
resource "aws_instance" "this" {
  ami           = data.aws_ami.this.id
  instance_type = var.instance_type
}

# Remote module call pinned to a specific version (TF_MOD_001), no provider
# block in the module (TF_MOD_002).
module "network" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.1.0"
}
