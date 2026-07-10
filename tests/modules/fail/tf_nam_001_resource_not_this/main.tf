# TF_NAM_001 KO: a resource whose type is unique in the module should be named
# "this" (or "these"), not "example".
resource "aws_s3_bucket" "example" {
  bucket = "my-example-bucket"
}
