# TG_DRY_002 KO: "region" is a literal used once, it should be inlined.
locals {
  region = "eu-west-3"
}

inputs = {
  region = local.region
}
