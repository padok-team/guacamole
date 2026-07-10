# TF_NAM_004 KO: a collection-typed variable should have a plural name.
variable "server" {
  description = "The list of servers to create"
  type        = list(string)
}
