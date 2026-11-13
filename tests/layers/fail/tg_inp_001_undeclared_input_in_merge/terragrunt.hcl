# TG_INP_001 KO: "unknown_input" given through merge() is not declared in the module.
terraform {
  source = "${get_terragrunt_dir()}/../../modules/bucket"
}

inputs = merge(
  { bucket_name = "my-app-bucket", region = "eu-west-3" },
  { unknown_input = "oops" },
)
