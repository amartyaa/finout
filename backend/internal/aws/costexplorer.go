package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	cetypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/finops/backend/internal/config"
)

type CostExplorerClient struct {
	cfg *config.Config
}

type DailyCostEntry struct {
	Date      string
	Service   string
	AccountID string
	Amount    float64
	Currency  string
}

func NewCostExplorerClient(cfg *config.Config) *CostExplorerClient {
	return &CostExplorerClient{cfg: cfg}
}

func (c *CostExplorerClient) FetchDailyCosts(ctx context.Context, roleARN, externalID string, startDate, endDate time.Time) ([]DailyCostEntry, error) {
	// First, assume the role
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(c.cfg.AWSRegion))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	stsClient := sts.NewFromConfig(awsCfg)
	assumeOutput, err := stsClient.AssumeRole(ctx, &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleARN),
		RoleSessionName: aws.String("finops-cost-sync"),
		ExternalId:      aws.String(externalID),
		DurationSeconds: aws.Int32(3600),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to assume role: %w", err)
	}

	// Create Cost Explorer client with assumed credentials
	creds := assumeOutput.Credentials
	ceCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("us-east-1"), // Cost Explorer is only in us-east-1
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			*creds.AccessKeyId,
			*creds.SecretAccessKey,
			*creds.SessionToken,
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CE config: %w", err)
	}

	ceClient := costexplorer.NewFromConfig(ceCfg)

	start := startDate.Format("2006-01-02")
	end := endDate.Format("2006-01-02")

	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &cetypes.DateInterval{
			Start: aws.String(start),
			End:   aws.String(end),
		},
		Granularity: cetypes.GranularityDaily,
		Metrics:     []string{"UnblendedCost"},
		GroupBy: []cetypes.GroupDefinition{
			{
				Type: cetypes.GroupDefinitionTypeDimension,
				Key:  aws.String("SERVICE"),
			},
		},
	}

	var entries []DailyCostEntry

	for {
		output, err := ceClient.GetCostAndUsage(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to get cost and usage: %w", err)
		}

		for _, result := range output.ResultsByTime {
			date := ""
			if result.TimePeriod != nil {
				date = *result.TimePeriod.Start
			}

			for _, group := range result.Groups {
				service := ""
				if len(group.Keys) > 0 {
					service = group.Keys[0]
				}

				amount := 0.0
				currency := "USD"
				if metric, ok := group.Metrics["UnblendedCost"]; ok {
					if metric.Amount != nil {
						fmt.Sscanf(*metric.Amount, "%f", &amount)
					}
					if metric.Unit != nil {
						currency = *metric.Unit
					}
				}

				entries = append(entries, DailyCostEntry{
					Date:     date,
					Service:  service,
					Amount:   amount,
					Currency: currency,
				})
			}
		}

		if output.NextPageToken == nil {
			break
		}
		input.NextPageToken = output.NextPageToken
	}

	return entries, nil
}
