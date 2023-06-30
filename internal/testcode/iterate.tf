locals {
  buckets = 3
}

provider "aws" {
  region = "us-east-1"
}

resource "aws_s3_bucket" "bucket_count" {
  count  = local.buckets
  bucket = "padok-my-tf-test-bucket"
}

resource "aws_s3_bucket" "bucket_foreach" {
  for_each = toset(["padok-my-tf-test-bucket-1", "padok-my-tf-test-bucket-2", "padok-my-tf-test-bucket-3"])
  bucket   = each.key
}

resource "aws_s3_bucket" "bucket_normal" {
  bucket = "padok-my-tf-test-bucket2"
}

module "bucket" {
  source = "./modules"
  name   = "padok-my-tf-test-bucket3"
}

module "bucket_count" {
  source = "./modules"
  count  = 3
  name   = "padok-my-tf-test-bucket3"
}

module "bucket_foreach" {
  source   = "./modules"
  for_each = toset(["padok-my-tf-test-bucket-5", "padok-my-tf-test-bucket-23", "padok-my-tf-test-bucket-43"])
  name     = "padok-my-tf-test-bucket3"
}
