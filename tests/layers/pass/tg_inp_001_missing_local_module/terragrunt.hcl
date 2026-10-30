# TG_INP_001 OK: the local module can't be loaded, the layer is skipped with a
# warning instead of aborting the whole run.
terraform {
  source = "../does_not_exist"
}

inputs = {
  name = "my-vpc"
}
