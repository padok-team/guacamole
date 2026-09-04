# TF_DAT_001 KO: the data source argument references a managed resource attribute
# (aws_instance.this.id) that is only known after apply, so the read is deferred
# and shows up as a potential change in the plan.
resource "aws_instance" "this" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t3.micro"
}

data "aws_instance" "this" {
  instance_id = aws_instance.this.id
}
