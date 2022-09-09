package gateway

import (
	"context"

	"github.com/botwayorg/railway-api/entity"
)

func (g *Gateway) GetStarters(ctx context.Context) ([]*entity.Starter, error) {
	gqlReq := g.NewRequestWithoutAuth(`
		query {
			getAllStarters {
				title
				url
				source
			}
		}
	`)

	var resp struct {
		Starters []*entity.Starter `json:"getAllStarters"`
	}

	if err := gqlReq.Run(ctx, &resp); err != nil {
		return nil, err
	}

	starters := resp.Starters

	return starters, nil
}
