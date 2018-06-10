package cluster

import (
	"context"

	"github.com/ehazlett/stellar"
	api "github.com/ehazlett/stellar/api/services/cluster/v1"
	nodeapi "github.com/ehazlett/stellar/api/services/node/v1"
)

func (s *service) Containers(ctx context.Context, req *api.ContainersRequest) (*api.ContainersResponse, error) {
	var containers []*nodeapi.Container

	resp, err := s.Nodes(ctx, &api.NodesRequest{})
	if err != nil {
		return nil, err
	}

	for _, node := range resp.Nodes {
		c, err := stellar.NewClient(node.Addr)
		if err != nil {
			return nil, err
		}
		cont, err := c.Containers()
		if err != nil {
			return nil, err
		}
		containers = append(containers, cont...)
		c.Close()
	}

	return &api.ContainersResponse{
		Containers: containers,
	}, nil
}