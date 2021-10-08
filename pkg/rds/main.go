package rds

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	awsTagAPI "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"k8s.io/kubectl/pkg/util/slice"
)

type Client struct {
	TagClient *awsTagAPI.Client
	Region    string
	RDSClient *rds.Client
	Config    aws.Config
}

type DBAuth struct {
	Username string
	Password string
	Host     string
	Port     int
	DBName   string
	Scheme   string
}

func New(region string, config aws.Config) *Client {
	return &Client{
		Region: region,
		Config: config,
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

type Meta struct {
	AllowedUsers []string `json:"allowed_users,omitempty"`
	Identifier   string   `json:"identifier"`
}

func (c Client) describeDatabase(ctx context.Context, identifier string) (*types.DBInstance, error) {
	out, err := c.RDSClient.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(identifier),
	})
	if err != nil {
		return nil, err
	}
	if len(out.DBInstances) == 0 {
		return nil, fmt.Errorf("database %q not found", identifier)
	}

	return &out.DBInstances[0], nil
}

func (c Client) GetDBCredentials(ctx context.Context, meta *Meta, user string) (*DBAuth, error) {
	if len(meta.AllowedUsers) > 0 {
		if !slice.ContainsString(meta.AllowedUsers, user, nil) {
			return nil, fmt.Errorf("RDS DB user must be one of %v", meta.AllowedUsers)
		}
	}

	db, err := c.describeDatabase(ctx, meta.Identifier)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s:%d", *db.Endpoint.Address, db.Endpoint.Port)

	token, err := auth.BuildAuthToken(ctx, endpoint, c.Region, user, c.Config.Credentials)
	if err != nil {
		return nil, err
	}

	return &DBAuth{
		Username: user,
		Password: token,
		Host:     *db.Endpoint.Address,
		Port:     int(db.Endpoint.Port),
		DBName:   *db.DBName,
		Scheme:   *db.Engine,
	}, nil
}
