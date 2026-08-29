package connections

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	log "github.com/Ptt-Alertor/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	dynamoDBReadyRetryInterval = 2 * time.Second
	dynamoDBTableRetryInterval = 250 * time.Millisecond
)

type dynamoDBAPI interface {
	ListTablesWithContext(aws.Context, *dynamodb.ListTablesInput, ...request.Option) (*dynamodb.ListTablesOutput, error)
	DescribeTableWithContext(aws.Context, *dynamodb.DescribeTableInput, ...request.Option) (*dynamodb.DescribeTableOutput, error)
	CreateTableWithContext(aws.Context, *dynamodb.CreateTableInput, ...request.Option) (*dynamodb.CreateTableOutput, error)
}

type tableDefinition struct {
	name string
	key  string
}

var requiredDynamoDBTables = []tableDefinition{
	{name: "boards", key: "Board"},
	{name: "articles", key: "Code"},
}

// EnsureDynamoDB waits for DynamoDB and idempotently creates the tables the
// application needs. Existing tables and their data are left untouched.
func EnsureDynamoDB(ctx context.Context) error {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}
	endpoint := os.Getenv("DB_CONNECTION")
	if endpoint == "" {
		return errors.New("DB_CONNECTION is not set")
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Endpoint:    aws.String(endpoint),
		Credentials: credentials.NewStaticCredentials("local", "local", ""),
	})
	if err != nil {
		return fmt.Errorf("create DynamoDB session: %w", err)
	}

	client := dynamodb.New(sess)
	if err := waitForDynamoDB(ctx, client, dynamoDBReadyRetryInterval); err != nil {
		return err
	}
	for _, table := range requiredDynamoDBTables {
		if err := ensureDynamoDBTable(ctx, client, table, dynamoDBTableRetryInterval); err != nil {
			return err
		}
	}
	return nil
}

func waitForDynamoDB(ctx context.Context, client dynamoDBAPI, retryInterval time.Duration) error {
	for {
		if _, err := client.ListTablesWithContext(ctx, &dynamodb.ListTablesInput{}); err == nil {
			log.Info("DynamoDB is ready")
			return nil
		} else {
			log.WithError(err).Warn("Waiting for DynamoDB")
		}

		if err := waitForRetry(ctx, retryInterval); err != nil {
			return fmt.Errorf("wait for DynamoDB: %w", err)
		}
	}
}

func ensureDynamoDBTable(ctx context.Context, client dynamoDBAPI, table tableDefinition, retryInterval time.Duration) error {
	input := &dynamodb.DescribeTableInput{TableName: aws.String(table.name)}
	if _, err := client.DescribeTableWithContext(ctx, input); err != nil {
		if !isAWSError(err, dynamodb.ErrCodeResourceNotFoundException) {
			return fmt.Errorf("describe DynamoDB table %s: %w", table.name, err)
		}

		_, err = client.CreateTableWithContext(ctx, &dynamodb.CreateTableInput{
			TableName: aws.String(table.name),
			AttributeDefinitions: []*dynamodb.AttributeDefinition{
				{
					AttributeName: aws.String(table.key),
					AttributeType: aws.String(dynamodb.ScalarAttributeTypeS),
				},
			},
			KeySchema: []*dynamodb.KeySchemaElement{
				{
					AttributeName: aws.String(table.key),
					KeyType:       aws.String(dynamodb.KeyTypeHash),
				},
			},
			BillingMode: aws.String(dynamodb.BillingModePayPerRequest),
		})
		if err != nil && !isAWSError(err, dynamodb.ErrCodeResourceInUseException) {
			return fmt.Errorf("create DynamoDB table %s: %w", table.name, err)
		}
		log.WithField("table", table.name).Info("DynamoDB table created")
	}

	for {
		output, err := client.DescribeTableWithContext(ctx, input)
		if err == nil && output.Table != nil && aws.StringValue(output.Table.TableStatus) == dynamodb.TableStatusActive {
			log.WithField("table", table.name).Info("DynamoDB table is ready")
			return nil
		}
		if err != nil && !isAWSError(err, dynamodb.ErrCodeResourceNotFoundException) {
			return fmt.Errorf("wait for DynamoDB table %s: %w", table.name, err)
		}
		if err := waitForRetry(ctx, retryInterval); err != nil {
			return fmt.Errorf("wait for DynamoDB table %s: %w", table.name, err)
		}
	}
}

func waitForRetry(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func isAWSError(err error, code string) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == code
}
