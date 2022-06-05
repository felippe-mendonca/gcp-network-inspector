package gcp

import (
	"context"
	"fmt"
	"sync"

	compute "cloud.google.com/go/compute/apiv1"
	"github.com/golang/protobuf/jsonpb"
	"golang.org/x/oauth2/google"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

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

	client, err := google.DefaultClient(ctx)
	if err != nil {
		return []*computepb.Subnetwork{}, err
	}
	defer client.CloseIdleConnections()

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

			res, err := client.Get(s)
			if err != nil {
				errChannel <- fmt.Errorf("gcp.ListSubnetworks: %v", err)
				return
			}

			if res.StatusCode != 200 {
				errChannel <- fmt.Errorf("gcp.ListSubnetworks: failed to get %s, StatusCode: %d", s, res.StatusCode)
				return
			}

			ss := &computepb.Subnetwork{}
			err = jsonpb.Unmarshal(res.Body, ss)
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
