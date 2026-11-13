# TG_DRY_002 OK: "owner" is a plain literal, but it is read by several layers
# below this file (dev and prod): it is reused.
locals {
  owner = "team-a"
}
