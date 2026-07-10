variable "instance_type" {
  description = "The type of EC2 instance to create"
  type        = string
  default     = "t3.micro"
}

variable "ami_name_pattern" {
  description = "The name pattern used to look up the AMI"
  type        = string
  default     = "amzn2-ami-hvm-*-x86_64-gp2"
}

# Collection type -> plural name (TF_NAM_004).
variable "subnets" {
  description = "The list of subnets to attach"
  type        = list(string)
  default     = []
}
