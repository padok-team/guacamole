provider "aws" {}

resource "random_pet" "this" {
  for_each = toset(var.pets)
  length   = 3
}

resource "random_pet" "random_pet_count" {
  count  = 130
  length = 4
}
