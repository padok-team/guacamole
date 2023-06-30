resource "random_pet" "this" {
  for_each = toset(var.pets)
  length   = 3
}
provider "aws" {
  
}

resource "random_pet" "random_pet_count" {
  count = 125
  length   = 3
}
