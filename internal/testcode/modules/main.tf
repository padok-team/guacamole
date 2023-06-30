resource "aws_s3_bucket" "this" {
  bucket = var.name
}

variable "name" {
}

# resource "aws_s3_bucket" "withcount" {
#   count  = 8
#   bucket = var.name
# }
