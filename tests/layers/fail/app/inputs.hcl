# TG_DRY_001 KO: "region" is already defined with the same value in common.hcl,
# so it is a duplicated input within the layer.
inputs = {
  region      = "eu-west-3"
  bucket_name = "my-app-bucket"
}
