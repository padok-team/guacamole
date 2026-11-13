# get_terragrunt_dir() evaluated from an included file is the including layer's dir.
terraform {
  source = "${get_terragrunt_dir()}/../../modules/bucket"
}
