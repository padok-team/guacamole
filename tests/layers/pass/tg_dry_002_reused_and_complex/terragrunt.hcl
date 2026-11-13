# TG_DRY_002 OK: "env" is reused, "tags" abstracts complexity (merge).
locals {
  env  = "dev"
  tags = merge({ team = "team-a" }, { env = local.env })
}

inputs = {
  name = "app-${local.env}"
  tags = local.tags
}
