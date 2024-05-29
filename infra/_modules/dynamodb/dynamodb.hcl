terraform {
  source = "../../../modules/dynamodb"
}

locals {
  config = read_terragrunt_config("../config.hcl")
}

inputs = {
  table_name = "${local.config.inputs.table_name_prefix}_${local.config.inputs.environment}"
  hash_key   = "instance_id"
  attributes = [
    {
      name = "instance_id"
      type = "S"
    },
    {
      name = "start_time"
      type = "S"
    },
    {
      name = "stop_time"
      type = "S"
    },
    {
      name = "timezone"
      type = "S"
    },
    {
      name = "aws_region"
      type = "S"
    },
    {
      name = "friendly_name"
      type = "S"
    }
  ]
  global_secondary_indexes = [
    {
      name            = "StartTimeIndex"
      hash_key        = "start_time"
      projection_type = "ALL"
    },
    {
      name            = "StopTimeIndex"
      hash_key        = "stop_time"
      projection_type = "ALL"
    },
    {
      name            = "TimezoneIndex"
      hash_key        = "timezone"
      projection_type = "ALL"
    },
    {
      name            = "AWSRegionIndex"
      hash_key        = "aws_region"
      projection_type = "ALL"
    },
    {
      name            = "FriendlyNameIndex"
      hash_key        = "friendly_name"
      projection_type = "ALL"
    }
  ]
}