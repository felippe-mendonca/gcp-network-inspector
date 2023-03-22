package network

import (
	"math"
	"net"
	"sort"

	"github.com/apparentlymart/go-cidr/cidr"
)

const (
	IPv4maskSize = 8 * net.IPv4len
)

func ipToInt(ip net.IP) (val int) {
	octet_base := 0x100000000
	for _, octet := range ip.To4() {
		octet_base >>= 8
		val += octet_base * int(octet)
	}
	return val
}

type Subnetworks []*net.IPNet

func (s Subnetworks) Len() int { return len(s) }

func (s Subnetworks) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s Subnetworks) Less(i, j int) bool {
	si := ipToInt(s[i].IP)
	sj := ipToInt(s[j].IP)
	return si < sj
}

func hostsInRange(begin net.IP, end net.IP) int {
	return ipToInt(end) - ipToInt(begin)
}

func onlySubnetworks(subnetworks Subnetworks, network *net.IPNet) (output Subnetworks) {
	for _, s := range subnetworks {
		begin, end := cidr.AddressRange(s)
		if network.Contains(begin) && network.Contains(end) {
			output = append(output, s)
		}
	}
	return output
}

func findSubnetwork(begin net.IP, hosts int) (subnet *net.IPNet, leftHosts int) {
	exp := int(math.Floor(math.Log2(float64(hosts))))
	netbits := IPv4maskSize - exp
	mask := net.CIDRMask(netbits, IPv4maskSize)
	subnet = &net.IPNet{IP: begin, Mask: mask}
	networkHosts := int(math.Pow(2, float64(exp)))
	leftHosts = hosts - networkHosts
	return subnet, leftHosts
}

func nextSubnetworkBegin(subnet *net.IPNet) net.IP {
	_, begin := cidr.AddressRange(subnet)
	begin = cidr.Inc(begin)
	return begin
}

func FindAvailableSubnetworks(subnetworks Subnetworks, network *net.IPNet) (availableSubNets Subnetworks, err error) {

	subnetworks = onlySubnetworks(subnetworks, network)

	if len(subnetworks) == 0 {
		return Subnetworks{network}, err
	}

	err = cidr.VerifyNoOverlap(subnetworks, network)
	if err != nil {
		return availableSubNets, err
	}

	sort.Sort(subnetworks)

	begin := network.IP
	for _, subnet := range subnetworks {
		end := subnet.IP
		hostsAvailable := hostsInRange(begin, end)

		for hostsAvailable > 0 {
			s, leftHosts := findSubnetwork(begin, hostsAvailable)
			availableSubNets = append(availableSubNets, s)
			begin = nextSubnetworkBegin(s)
			hostsAvailable = leftHosts
		}

		begin = nextSubnetworkBegin(subnet)
	}

	return availableSubNets, err
}
