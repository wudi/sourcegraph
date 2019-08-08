package rules

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/campaigns"
)

func init() {
	graphqlbackend.Rules = GraphQLResolver{}
	graphqlbackend.Campaigns = campaigns.GraphQLResolver{}
}

type mockCampaign struct {
	graphqlbackend.Campaign
	id int64
}

func (v mockCampaign) Rules(ctx context.Context, arg *graphqlutil.ConnectionArgs) (graphqlbackend.RuleConnection, error) {
	return graphqlbackend.RulesInRuleContainer(ctx, graphqlbackend.MarshalCampaignID(v.id), arg)
}