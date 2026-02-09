package awssts

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Client is the minimal interface we need from STS.
// Keeping it small makes unit testing easy.
type Client interface {
	GetSessionToken(ctx context.Context, in GetSessionTokenInput) (GetSessionTokenOutput, error)
}

type GetSessionTokenInput struct {
	SerialNumber    string
	TokenCode       string
	DurationSeconds int32
}

type GetSessionTokenOutput struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Expiration      time.Time
}

// RealClient calls AWS STS using AWS SDK for Go v2.
type RealClient struct {
	api *sts.Client
}

// NewRealClient constructs an STS client that authenticates using the provided
// long-term credentials. Region is required by the AWS SDK; STS is a global
// service but still expects a region to be set (we default in higher layers).
func NewRealClient(ctx context.Context, region, accessKeyID, secretAccessKey string) (*RealClient, error) {
	if region == "" {
		return nil, fmt.Errorf("region is empty")
	}
	if accessKeyID == "" || secretAccessKey == "" {
		return nil, fmt.Errorf("access key id/secret access key must be set")
	}

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	return &RealClient{api: sts.NewFromConfig(cfg)}, nil
}

func (c *RealClient) GetSessionToken(ctx context.Context, in GetSessionTokenInput) (GetSessionTokenOutput, error) {
	out, err := c.api.GetSessionToken(ctx, &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int32(in.DurationSeconds),
		SerialNumber:    aws.String(in.SerialNumber),
		TokenCode:       aws.String(in.TokenCode),
	})
	if err != nil {
		return GetSessionTokenOutput{}, fmt.Errorf("sts get-session-token: %w", err)
	}
	if out.Credentials == nil {
		return GetSessionTokenOutput{}, fmt.Errorf("sts get-session-token: no credentials in response")
	}
	return GetSessionTokenOutput{
		AccessKeyID:     aws.ToString(out.Credentials.AccessKeyId),
		SecretAccessKey: aws.ToString(out.Credentials.SecretAccessKey),
		SessionToken:    aws.ToString(out.Credentials.SessionToken),
		Expiration:      aws.ToTime(out.Credentials.Expiration).UTC(),
	}, nil
}
