# TF_NAM_003 KO: resource names should not stutter with their type (both names
# repeat the "bucket" word of "aws_s3_bucket").
# Two resources of the same type so this fixture does not also trip TF_NAM_001.
resource "aws_s3_bucket" "logs_bucket" {
  bucket = "example-logs"
}

resource "aws_s3_bucket" "data_bucket" {
  bucket = "example-data"
}
