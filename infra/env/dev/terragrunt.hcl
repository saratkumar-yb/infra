locals {
  config = read_terragrunt_config("${get_parent_terragrunt_dir()}/config.hcl")
}

remote_state {
  backend = "s3"
  generate = {
    path      = "backend.tf"
    if_exists = "overwrite"
  }
  config = {
    bucket         = "yb-infra-terraform-state"
    key            = "yb_infra/${path_relative_to_include()}/terraform.tfstate"
    region         = "us-east-1"
    profile        = "${local.config.inputs.environment}"
    encrypt        = true
    dynamodb_table = "${local.config.inputs.environment}-lock-table"
  }
}

generate "versions" {
  path      = "versions_override.tf"
  if_exists = "overwrite_terragrunt"
  contents  = <<EOF
    terraform {
      required_providers {
        aws = {
          source  = "hashicorp/aws"
          version = "4.66.1"
        }
      }
    }
EOF
}

generate "provider" {
  path = "provider.tf"
  if_exists = "overwrite_terragrunt"
  contents = <<EOF
provider "aws" {
  region                      = "${local.config.inputs.aws_region}"
  profile                     = "${local.config.inputs.environment}"
}
EOF
}