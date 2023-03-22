package network

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ipToInt(t *testing.T) {
	tests := []struct {
		IP    net.IP
		IPInt int
	}{
		{
			IP:    net.ParseIP("10.10.10.10").To4(),
			IPInt: 168430090,
		},
		{
			IP:    net.ParseIP("1.2.3.4").To4(),
			IPInt: 16909060,
		},
		{
			IP:    net.ParseIP("0.0.0.0").To4(),
			IPInt: 0,
		},
	}

	for _, test := range tests {
		assert := assert.New(t)
		ipInt := ipToInt(test.IP)
		assert.Equal(test.IPInt, ipInt)
	}
}

func Test_FindSubnetwork(t *testing.T) {
	// https://www.davidc.net/sites/default/subnets/subnets.html?network=10.0.0.0&mask=8&division=9.350
	tests := []struct {
		Begin     net.IP
		Hosts     int
		SubNet    *net.IPNet
		LeftHosts int
	}{
		{
			Begin:     net.ParseIP("10.0.0.0").To4(),
			Hosts:     4194304,
			SubNet:    &net.IPNet{IP: net.ParseIP("10.0.0.0").To4(), Mask: net.CIDRMask(10, IPv4maskSize)},
			LeftHosts: 0,
		},
		{
			Begin:     net.ParseIP("10.128.0.0").To4(),
			Hosts:     4194304 + 2097152,
			SubNet:    &net.IPNet{IP: net.ParseIP("10.128.0.0").To4(), Mask: net.CIDRMask(10, IPv4maskSize)},
			LeftHosts: 2097152,
		},
	}

	for _, test := range tests {
		assert := assert.New(t)
		subnet, leftHosts := findSubnetwork(test.Begin, test.Hosts)
		assert.Equal(test.SubNet, subnet)
		assert.Equal(test.LeftHosts, leftHosts)
	}

}

func Test_FindAvailableSubnetworks(t *testing.T) {
	tests := []struct {
		Description      string
		Network          *net.IPNet
		SubNets          Subnetworks
		AvailableSubNets Subnetworks
		Err              error
	}{
		{
			// https://www.davidc.net/sites/default/subnets/subnets.html?network=10.0.0.0&mask=8&division=11.b21
			Description: "First range available, intermediate ranges, last range not available.",
			Network:     &net.IPNet{IP: net.ParseIP("10.0.0.0").To4(), Mask: net.CIDRMask(8, IPv4maskSize)},
			SubNets: Subnetworks{
				&net.IPNet{IP: net.ParseIP("10.64.0.0").To4(), Mask: net.CIDRMask(11, IPv4maskSize)},
				&net.IPNet{IP: net.ParseIP("10.96.0.0").To4(), Mask: net.CIDRMask(12, IPv4maskSize)},
				&net.IPNet{IP: net.ParseIP("10.112.0.0").To4(), Mask: net.CIDRMask(12, IPv4maskSize)},
				&net.IPNet{IP: net.ParseIP("10.192.0.0").To4(), Mask: net.CIDRMask(10, IPv4maskSize)},
			},
			AvailableSubNets: Subnetworks{
				&net.IPNet{IP: net.ParseIP("10.0.0.0").To4(), Mask: net.CIDRMask(10, IPv4maskSize)},
				&net.IPNet{IP: net.ParseIP("10.128.0.0").To4(), Mask: net.CIDRMask(10, IPv4maskSize)},
			},
			Err: nil,
		},
		{
			// https://www.davidc.net/sites/default/subnets/subnets.html?network=10.0.0.0&mask=8&division=7.31
			Description: "First and last range not available, subnetwork out of network",
			Network:     &net.IPNet{IP: net.ParseIP("10.0.0.0").To4(), Mask: net.CIDRMask(8, IPv4maskSize)},
			SubNets: Subnetworks{
				&net.IPNet{IP: net.ParseIP("10.0.0.0").To4(), Mask: net.CIDRMask(10, IPv4maskSize)},
				&net.IPNet{IP: net.ParseIP("10.192.0.0").To4(), Mask: net.CIDRMask(10, IPv4maskSize)},
				&net.IPNet{IP: net.ParseIP("172.31.0.0").To4(), Mask: net.CIDRMask(16, IPv4maskSize)},
			},
			AvailableSubNets: Subnetworks{
				&net.IPNet{IP: net.ParseIP("10.64.0.0").To4(), Mask: net.CIDRMask(9, IPv4maskSize)},
			},
			Err: nil,
		},
		{
			Description: "All range available",
			Network:     &net.IPNet{IP: net.ParseIP("10.0.0.0").To4(), Mask: net.CIDRMask(8, IPv4maskSize)},
			SubNets:     Subnetworks{},
			AvailableSubNets: Subnetworks{
				&net.IPNet{IP: net.ParseIP("10.0.0.0").To4(), Mask: net.CIDRMask(8, IPv4maskSize)},
			},
			Err: nil,
		},
		{
			Description: "SubNets with overlap",
			Network:     &net.IPNet{IP: net.ParseIP("10.0.0.0").To4(), Mask: net.CIDRMask(8, IPv4maskSize)},
			SubNets: Subnetworks{
				&net.IPNet{IP: net.ParseIP("10.64.0.0").To4(), Mask: net.CIDRMask(10, IPv4maskSize)},
				&net.IPNet{IP: net.ParseIP("10.64.0.0").To4(), Mask: net.CIDRMask(11, IPv4maskSize)},
			},
			AvailableSubNets: Subnetworks{},
			Err:              fmt.Errorf("%s overlaps with %s", "10.64.0.0/11", "10.64.0.0/10"),
		},
		{
			Description: "Multiple available subnetworks in sequence",
			Network:     &net.IPNet{IP: net.ParseIP("10.0.0.0").To4(), Mask: net.CIDRMask(8, IPv4maskSize)},
			SubNets: Subnetworks{
				&net.IPNet{IP: net.ParseIP("10.64.0.0").To4(), Mask: net.CIDRMask(11, IPv4maskSize)},
				&net.IPNet{IP: net.ParseIP("10.112.0.0").To4(), Mask: net.CIDRMask(12, IPv4maskSize)},
				&net.IPNet{IP: net.ParseIP("10.136.0.0").To4(), Mask: net.CIDRMask(13, IPv4maskSize)},
				&net.IPNet{IP: net.ParseIP("10.160.0.0").To4(), Mask: net.CIDRMask(11, IPv4maskSize)},
				&net.IPNet{IP: net.ParseIP("10.192.0.0").To4(), Mask: net.CIDRMask(11, IPv4maskSize)},
				&net.IPNet{IP: net.ParseIP("10.248.0.0").To4(), Mask: net.CIDRMask(14, IPv4maskSize)},
				&net.IPNet{IP: net.ParseIP("10.252.0.0").To4(), Mask: net.CIDRMask(15, IPv4maskSize)},
				&net.IPNet{IP: net.ParseIP("10.255.0.0").To4(), Mask: net.CIDRMask(16, IPv4maskSize)},
			},
			AvailableSubNets: Subnetworks{
				&net.IPNet{IP: net.ParseIP("10.0.0.0").To4(), Mask: net.CIDRMask(10, IPv4maskSize)},
				&net.IPNet{IP: net.ParseIP("10.96.0.0").To4(), Mask: net.CIDRMask(12, IPv4maskSize)},
				&net.IPNet{IP: net.ParseIP("10.128.0.0").To4(), Mask: net.CIDRMask(13, IPv4maskSize)},
				&net.IPNet{IP: net.ParseIP("10.144.0.0").To4(), Mask: net.CIDRMask(12, IPv4maskSize)},
				&net.IPNet{IP: net.ParseIP("10.224.0.0").To4(), Mask: net.CIDRMask(12, IPv4maskSize)},
				&net.IPNet{IP: net.ParseIP("10.240.0.0").To4(), Mask: net.CIDRMask(13, IPv4maskSize)},
				&net.IPNet{IP: net.ParseIP("10.254.0.0").To4(), Mask: net.CIDRMask(16, IPv4maskSize)},
			},
			Err: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Description, func(t *testing.T) {
			assert := assert.New(t)
			availableSubNets, err := FindAvailableSubnetworks(test.SubNets, test.Network)
			assert.ElementsMatch(test.AvailableSubNets, availableSubNets)
			assert.Equal(test.Err, err)
		})
	}
}
