include "root" {
  path = find_in_parent_folders()
}

locals {
  config = read_terragrunt_config("../config.hcl")
}

include "env" {
  path = "${get_repo_root()}/infra/_modules/dynamodb/dynamodb.hcl"
}