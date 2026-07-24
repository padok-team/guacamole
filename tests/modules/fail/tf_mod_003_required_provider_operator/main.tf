# TF_MOD_003 KO: required provider version should use the "~>" operator.
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
  }
}
