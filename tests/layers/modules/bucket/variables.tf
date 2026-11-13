variable "bucket_name" {
  description = "Name of the bucket."
  type        = string
}

variable "region" {
  description = "Region of the bucket."
  type        = string
}

# Optional on purpose: layers are free to leave it unset.
variable "tags" {
  description = "Tags of the bucket."
  type        = map(string)
  default     = {}
}
