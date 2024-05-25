variable "table_name" {
  description = "DynamoDB table name"
  type        = string
}

variable "hash_key" {
  description = "DynamoDB table hash key"
  type        = string
}

variable "attributes" {
  description = "DynamoDB table attributes"
  type = list(object({
    name = string
    type = string
  }))
}

variable "global_secondary_indexes" {
  description = "DynamoDB table global secondary indexes"
  type = list(object({
    name               = string
    hash_key           = string
    projection_type    = string
  }))
  default = []
}