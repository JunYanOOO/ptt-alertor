package connections

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type fakeDynamoDB struct {
	tables       map[string]bool
	listFailures int
	created      []string
}

func (fake *fakeDynamoDB) ListTablesWithContext(_ aws.Context, _ *dynamodb.ListTablesInput, _ ...request.Option) (*dynamodb.ListTablesOutput, error) {
	if fake.listFailures > 0 {
		fake.listFailures--
		return nil, awserr.New("Unavailable", "not ready", nil)
	}
	return &dynamodb.ListTablesOutput{}, nil
}

func (fake *fakeDynamoDB) DescribeTableWithContext(_ aws.Context, input *dynamodb.DescribeTableInput, _ ...request.Option) (*dynamodb.DescribeTableOutput, error) {
	if !fake.tables[aws.StringValue(input.TableName)] {
		return nil, awserr.New(dynamodb.ErrCodeResourceNotFoundException, "missing", nil)
	}
	return &dynamodb.DescribeTableOutput{
		Table: &dynamodb.TableDescription{TableStatus: aws.String(dynamodb.TableStatusActive)},
	}, nil
}

func (fake *fakeDynamoDB) CreateTableWithContext(_ aws.Context, input *dynamodb.CreateTableInput, _ ...request.Option) (*dynamodb.CreateTableOutput, error) {
	name := aws.StringValue(input.TableName)
	fake.tables[name] = true
	fake.created = append(fake.created, name)
	return &dynamodb.CreateTableOutput{}, nil
}

func TestWaitForDynamoDBRetries(t *testing.T) {
	fake := &fakeDynamoDB{tables: map[string]bool{}, listFailures: 2}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := waitForDynamoDB(ctx, fake, time.Millisecond); err != nil {
		t.Fatalf("waitForDynamoDB() error = %v", err)
	}
	if fake.listFailures != 0 {
		t.Fatalf("waitForDynamoDB() did not retry all failures")
	}
}

func TestEnsureDynamoDBTableCreatesMissingTable(t *testing.T) {
	fake := &fakeDynamoDB{tables: map[string]bool{}}
	table := tableDefinition{name: "boards", key: "Board"}

	if err := ensureDynamoDBTable(context.Background(), fake, table, time.Millisecond); err != nil {
		t.Fatalf("ensureDynamoDBTable() error = %v", err)
	}
	if len(fake.created) != 1 || fake.created[0] != table.name {
		t.Fatalf("created tables = %v, want [%s]", fake.created, table.name)
	}
}

func TestEnsureDynamoDBTableKeepsExistingTable(t *testing.T) {
	fake := &fakeDynamoDB{tables: map[string]bool{"articles": true}}
	table := tableDefinition{name: "articles", key: "Code"}

	if err := ensureDynamoDBTable(context.Background(), fake, table, time.Millisecond); err != nil {
		t.Fatalf("ensureDynamoDBTable() error = %v", err)
	}
	if len(fake.created) != 0 {
		t.Fatalf("created tables = %v, want none", fake.created)
	}
}
