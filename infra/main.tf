locals {
  common_labels = {
    "created_by"  = "mo"
    "created_for" = "mo-try-out-pubsub-locally"
  }
  key                        = var.unique_key
  message_retention_duration = "86400s" # 1 day
}
