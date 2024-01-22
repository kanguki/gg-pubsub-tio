variable "project_name" {
  description = "project name"
}

variable "project_region" {
  description = "project region"
  default     = "us-central1"
}

variable "project_zone" {
  description = "project region"
  default     = "us-central1-b"
}

variable "unique_key" {
  description = "a unique key to create resources in this project"
  default     = "mo-pip1762"
}
