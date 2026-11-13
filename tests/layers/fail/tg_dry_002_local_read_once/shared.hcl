# TG_DRY_002 KO: "owner" is a plain literal read by a single layer (dev) only,
# so it is not reused: inline it in dev/terragrunt.hcl.
locals {
  owner = "team-a"
}
