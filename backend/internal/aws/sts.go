package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/finops/backend/internal/config"
)

type STSClient struct {
	cfg *config.Config
}

func NewSTSClient(cfg *config.Config) *STSClient {
	return &STSClient{cfg: cfg}
}

func (s *STSClient) ValidateRole(ctx context.Context, roleARN, externalID string) error {
	_, err := s.AssumeRole(ctx, roleARN, externalID)
	return err
}

func (s *STSClient) AssumeRole(ctx context.Context, roleARN, externalID string) (*sts.AssumeRoleOutput, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(s.cfg.AWSRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := sts.NewFromConfig(awsCfg)

	output, err := client.AssumeRole(ctx, &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleARN),
		RoleSessionName: aws.String("finops-saas-session"),
		ExternalId:      aws.String(externalID),
		DurationSeconds: aws.Int32(3600),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to assume role %s: %w", roleARN, err)
	}

	return output, nil
}
