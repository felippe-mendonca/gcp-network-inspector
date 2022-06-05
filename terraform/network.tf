locals {
  subnetworks = [
    {
      name : "subnet-01",
      cidr : "10.64.0.0/11",
      secondary_ip_range : [
        {
          name : "range-01",
          cidr : "10.112.0.0/12",
        }
      ]
    },
    {
      name : "subnet-02",
      cidr : "10.136.0.0/13",
      secondary_ip_range : []
    },
    {
      name : "subnet-03",
      cidr : "10.160.0.0/11",
      secondary_ip_range : [
        {
          name : "range-01",
          cidr : "10.192.0.0/11",
        }
      ]
    },
    {
      name : "subnet-05",
      cidr : "10.248.0.0/14",
      secondary_ip_range : []
    },
    {
      name : "subnet-06",
      cidr : "10.252.0.0/15",
      secondary_ip_range : []
    },
    {
      name : "subnet-07",
      cidr : "10.255.0.0/16",
      secondary_ip_range : []
    },
  ]
}

resource "google_compute_network" "vpc_network" {
  name                    = "vpc-network"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "subnetworks" {
  for_each = {
    for _, v in local.subnetworks : v.name => v
  }

  name          = each.value.name
  ip_cidr_range = each.value.cidr
  region        = var.region
  network       = google_compute_network.vpc_network.id

  dynamic "secondary_ip_range" {
    for_each = { for _, vv in each.value.secondary_ip_range : vv.name => vv }
    content {
      range_name    = secondary_ip_range.value.name
      ip_cidr_range = secondary_ip_range.value.cidr
    }
  }
}
