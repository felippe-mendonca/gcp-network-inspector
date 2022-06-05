terraform {
  required_version = ">=1.0"

  required_providers {
    google = {
      version = ">=4.23.0"
      source  = "hashicorp/google"
    }
  }
  backend "gcs" {
    bucket = "my-tf-states"
    prefix = "terraform/states/gcp-network-inspector"
  }
}

provider "google" {
  project = var.gcp_project
}
