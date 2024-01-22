resource "google_pubsub_topic" "mo" {
  name                       = local.key
  labels                     = merge({}, local.common_labels)
  message_retention_duration = local.message_retention_duration
}

resource "google_pubsub_subscription" "mo_main" {
  name                       = local.key
  topic                      = google_pubsub_topic.mo.name
  ack_deadline_seconds       = 60
  labels                     = merge({}, local.common_labels)
  message_retention_duration = local.message_retention_duration
  retain_acked_messages      = true
  enable_message_ordering    = false
}

resource "google_service_account" "mo" {
  account_id   = local.key
  display_name = local.key
}

resource "google_pubsub_subscription_iam_binding" "mo_editor" {
  subscription = google_pubsub_subscription.mo_main.name
  role         = "roles/editor"
  members = [
    "serviceAccount:${google_service_account.mo.email}"
  ]
}

resource "google_pubsub_topic_iam_binding" "mo_editor" {
  topic = google_pubsub_topic.mo.name
  role  = "roles/editor"
  members = [
    "serviceAccount:${google_service_account.mo.email}"
  ]
}

resource "google_service_account_key" "mo" {
  service_account_id = google_service_account.mo.account_id
}

resource "local_file" "mo_sa_key" {
  content  = base64decode(google_service_account_key.mo.private_key)
  filename = "credentials.json"
}
