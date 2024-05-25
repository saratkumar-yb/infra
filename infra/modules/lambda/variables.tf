variable "function_name" {
  description = "Lambda function name"
  type        = string
}

variable "handler" {
  description = "Lambda function handler"
  type        = string
}

variable "runtime" {
  description = "Lambda function runtime"
  type        = string
}

variable "memory_size" {
  description = "Lambda function memory size"
  type        = number
  default     = 128
}

variable "timeout" {
  description = "Lambda function timeout"
  type        = number
  default     = 30
}

variable "environment" {
  description = "Lambda environment variables"
  type = object({
    variables = map(string)
  })
  default = {
    variables = {}
  }
}

variable "zip_file" {
  description = "Path to the Lambda function zip file"
  type        = string
}
