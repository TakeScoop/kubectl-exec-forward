package rds

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdsTypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	awsTagAPI "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	tagTypes "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
)

type Client struct {
	TagClient *awsTagAPI.Client
	RDSClient *rds.Client
}

func New(region string, config aws.Config) *Client {
	return &Client{
		TagClient: awsTagAPI.New(awsTagAPI.Options{
			Region:      region,
			Credentials: config.Credentials,
		}),
		RDSClient: rds.New(rds.Options{
			Region:      region,
			Credentials: config.Credentials,
		}),
	}
}

func (c Client) GetDBInstanceByTags(ctx context.Context, tagFilters []tagTypes.TagFilter) (*rdsTypes.DBInstance, error) {
	resources, err := c.TagClient.GetResources(ctx, &awsTagAPI.GetResourcesInput{
		ResourceTypeFilters: []string{
			"rds:db",
		},
		TagFilters: tagFilters,
	})
	if err != nil {
		return nil, err
	}
	if len(resources.ResourceTagMappingList) == 0 {
		return nil, fmt.Errorf("no database found")
	}
	s := *resources.ResourceTagMappingList[0].ResourceARN
	identifier := s[strings.LastIndex(s, ":")+1:]

	instances, err := c.RDSClient.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: &identifier,
	})
	if err != nil {
		return nil, err
	}
	if len(instances.DBInstances) == 0 {
		return nil, fmt.Errorf("could not find database with identifier %q", identifier)
	}

	return &instances.DBInstances[0], nil
}

func (c Client) GetAuthToken(ctx context.Context, instance *rdsTypes.DBInstance, region string, dbUser string, config aws.Config) (string, error) {
	endpoint := fmt.Sprintf("%s:%d", *instance.Endpoint.Address, instance.Endpoint.Port)

	return auth.BuildAuthToken(ctx, endpoint, region, dbUser, config.Credentials)
}
