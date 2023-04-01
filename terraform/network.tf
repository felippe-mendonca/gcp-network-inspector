locals {
  // https://www.davidc.net/sites/default/subnets/subnets.html?network=10.0.0.0&mask=8&division=29.b2f05550

  subnetworks = [
    {
      name : "subnet-01",
      cidr : "10.0.0.0/10",
      secondary_ip_range : [
        {
          name : "range-01",
          cidr : "10.96.0.0/12",
        }
      ]
    },
    {
      name : "subnet-02",
      cidr : "10.128.0.0/13",
      secondary_ip_range : []
    },
    {
      name : "subnet-03",
      cidr : "10.144.0.0/12",
      secondary_ip_range : [
        {
          name : "range-01",
          cidr : "10.224.0.0/12",
        }
      ]
    },
    {
      name : "subnet-05",
      cidr : "10.240.0.0/13",
      secondary_ip_range : []
    },
    {
      name : "subnet-06",
      cidr : "10.254.0.0/16",
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
