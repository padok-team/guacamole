# TF_NAM_002 KO: resource names should be snake_case, not camelCase.
# Two resources of the same type so this fixture does not also trip TF_NAM_001
# (which only targets types that are unique within the module).
resource "aws_s3_bucket" "myBucket" {
  bucket = "my-bucket"
}

resource "aws_s3_bucket" "otherBucket" {
  bucket = "other-bucket"
}
