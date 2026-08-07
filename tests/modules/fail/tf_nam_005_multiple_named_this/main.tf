# TF_NAM_005 KO: when several resources share a type, they must not be named
# "this" or "these".
resource "aws_s3_bucket" "this" {
  bucket = "first-bucket"
}

resource "aws_s3_bucket" "these" {
  bucket = "second-bucket"
}
