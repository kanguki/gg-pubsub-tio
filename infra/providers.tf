terraform {
  required_version = ">= 1.6.6"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.10"
    }
  }
}

provider "google" {
  project = var.project_name
  region  = var.project_region
  zone    = var.project_zone
}
