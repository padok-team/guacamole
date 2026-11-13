# TG_INP_001 KO: "unknown_input" is not declared in the module set by module.hcl.
inputs = {
  bucket_name   = "my-app-bucket"
  region        = "eu-west-3"
  unknown_input = "oops"
}
