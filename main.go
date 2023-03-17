package main

import (
	"context"
	"fmt"
	"net"

	"gcp-network-inspector/pkg/gcp"
	"gcp-network-inspector/pkg/network"
)

func main() {
	project_name := "k8s-playground-123456"
	network_name := "vpc-network"

	ctx := context.Background()

	gcpNetwork, err := gcp.GetNetwork(ctx, network_name, project_name)
	if err != nil {
		panic(err)
	}

	_, networkCidr, err := net.ParseCIDR("10.0.0.0/8")
	if err != nil {
		panic(err)
	}

	gcpSubnets, err := gcp.ListSubnetworks(ctx, gcpNetwork)
	if err != nil {
		panic(err)
	}

	subnetworks := make([]*net.IPNet, 0)
	for _, gcpSubnet := range gcpSubnets {
		if _, s, err := net.ParseCIDR(*gcpSubnet.IpCidrRange); err == nil {
			subnetworks = append(subnetworks, s)
		}
	}

	for _, s := range subnetworks {
		fmt.Println(s)
	}

	fmt.Println("-----------------")
	availableSubnetworks, _ := network.FindAvailableSubnetworks(subnetworks, networkCidr)
	for _, s := range availableSubnetworks {
		fmt.Println(s)
	}

}
