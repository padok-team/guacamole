# TF_VAR_002 KO: a variable should declare a specific type, not "any".
variable "configuration" {
  description = "The configuration object"
  type        = any
}
