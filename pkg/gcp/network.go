package gcp

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	compute "cloud.google.com/go/compute/apiv1"
	"github.com/golang/protobuf/jsonpb"
	"golang.org/x/oauth2/google"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

type SubnetworkClient struct {
	Client *http.Client
}

func NewSubnetworkClient(ctx context.Context) (*SubnetworkClient, error) {
	client, err := google.DefaultClient(ctx)
	if err != nil {
		return &SubnetworkClient{}, fmt.Errorf("gcp.NewSubnetworkClient: %v", err)

	}
	return &SubnetworkClient{Client: client}, err
}

func (sc *SubnetworkClient) GetSubnetwork(subnet string) (subnetPb *computepb.Subnetwork, err error) {
	res, err := sc.Client.Get(subnet)
	if err != nil {
		return subnetPb, fmt.Errorf("gcp.SubnetworkClient.GetSubnetwork: %v", err)
	}

	if res.StatusCode != 200 {
		return subnetPb, fmt.Errorf("gcp.SubnetworkClient.GetSubnetwork: failed to get %s, StatusCode: %d", subnet, res.StatusCode)
	}

	subnetPb = &computepb.Subnetwork{}
	err = jsonpb.Unmarshal(res.Body, subnetPb)
	if err != nil {
		return subnetPb, fmt.Errorf("gcp.SubnetworkClient.GetSubnetwork: %v", err)
	}

	return subnetPb, err
}

func GetNetwork(ctx context.Context, networkName, projectName string) (*computepb.Network, error) {
	networkClient, err := compute.NewNetworksRESTClient(ctx)
	if err != nil {
		return &computepb.Network{}, fmt.Errorf("gcp.GetNetwork: %v", err)
	}
	defer networkClient.Close()

	req := &computepb.GetNetworkRequest{
		Project: projectName,
		Network: networkName,
	}

	net, err := networkClient.Get(ctx, req)
	if err != nil {
		return &computepb.Network{}, fmt.Errorf("gcp.GetNetwork: %v", err)
	}

	return net, nil
}

func ListSubnetworks(ctx context.Context, network *computepb.Network) (subnetworks []*computepb.Subnetwork, err error) {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	client, err := NewSubnetworkClient(ctx)
	if err != nil {
		return subnetworks, err
	}

	var wg sync.WaitGroup
	var mutex sync.Mutex
	errChannel := make(chan error)
	doneChannel := make(chan bool)

	for _, s := range network.GetSubnetworks() {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
			}

			ss, err := client.GetSubnetwork(s)
			if err != nil {
				errChannel <- fmt.Errorf("gcp.ListSubnetworks: %v", err)
				return
			}

			mutex.Lock()
			defer mutex.Unlock()
			subnetworks = append(subnetworks, ss)
		}(s)
	}

	go func() {
		wg.Wait()
		doneChannel <- true
	}()

	select {
	case <-doneChannel:
		return subnetworks, err
	case err = <-errChannel:
		return []*computepb.Subnetwork{}, err
	}
}
