terraform {
  source = "../../../modules/lambda"
}

locals {
  config = read_terragrunt_config("../config.hcl")
}

inputs = {
  function_name  = "instance-scheduler-lambda-${local.config.inputs.environment}"
  handler        = "main"
  runtime        = "provided.al2"
  memory_size    = 128
  timeout        = 30
  environment = {
    variables = {
      TABLE_NAME = "${local.config.inputs.table_name_prefix}_${local.config.inputs.environment}"
    }
  }
  zip_file       = "${get_repo_root()}/lambda/yb_infra.zip"
}