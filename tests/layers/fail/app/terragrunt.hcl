include "root" {
  path           = find_in_parent_folders("root.hcl")
  merge_strategy = "deep"
}

include "common" {
  path           = "common.hcl"
  merge_strategy = "deep"
}

include "inputs" {
  path           = "inputs.hcl"
  merge_strategy = "deep"
}
