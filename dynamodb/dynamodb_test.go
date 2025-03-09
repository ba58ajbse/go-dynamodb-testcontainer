package dynamodb_test

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	mydynamo "godynamodb/dynamodb"
	"godynamodb/testutil"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
)

var endpoint string

func TestMain(m *testing.M) {
	localStackEndpoint, cleanup := testutil.SetupLocalStack(context.TODO())
	endpoint = localStackEndpoint

	code := m.Run()

	cleanup()

	os.Exit(code)
}

func TestNewDynamoDB(t *testing.T) {
	t.Parallel()
	type args struct {
		endpoint string
		table    string
	}
	cases := map[string]struct {
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		"case 1": {
			args: args{
				endpoint: endpoint,
				table:    "Session",
			},
		},
		"case 2": {
			args: args{
				endpoint: endpoint,
				table:    "",
			},
			wantErr: true,
		},
	}
	for testName, tt := range cases {
		t.Run(testName, func(t *testing.T) {
			got, err := mydynamo.NewDynamoDB(tt.args.endpoint, tt.args.table)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDynamoDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				assert.NotNil(t, got.Client)
				assert.Equal(t, got.Table, tt.args.table)
			}
		})
	}
}

func TestDynamoDB_SaveSession(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	dynamo := setupDynamoTest(ctx, t, "Session_SaveSession")

	cases := map[string]struct {
		args    mydynamo.Session
		wantErr bool
	}{
		"case 1": {
			args: mydynamo.Session{
				ID:        "1234",
				SessionID: "abcde",
				CreatedAt: "2025-02-01 12:34:56",
				Expire:    "2025-02-01 13:34:56",
				TTL:       1738384496,
			},
			wantErr: false,
		},
		"case 2": {
			args: mydynamo.Session{
				ID:        "",
				SessionID: "abcde",
				CreatedAt: "2025-02-01 12:34:56",
				Expire:    "2025-02-01 13:34:56",
				TTL:       1738384496,
			},
			wantErr: true,
		},
		"case 3": {
			args: mydynamo.Session{
				ID:        "1234",
				SessionID: "",
				CreatedAt: "2025-02-01 12:34:56",
				Expire:    "2025-02-01 13:34:56",
				TTL:       1738384496,
			},
			wantErr: true,
		},
	}
	for testName, tt := range cases {
		t.Run(testName, func(t *testing.T) {
			if err := dynamo.SaveSession(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("DynamoDB.SaveSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDynamoDB_GetSession(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	session := mydynamo.Session{
		ID:        "1234",
		SessionID: "abcde",
		CreatedAt: "2025-02-01 12:34:56",
		Expire:    "2025-02-01 13:34:56",
		TTL:       1738384496,
	}

	dynamo := setupDynamoTest(ctx, t, "Session_GetSession")
	saveTestItem(t, dynamo, session)

	type args struct {
		ID        string
		sessionID string
		now       string
	}
	cases := map[string]struct {
		args    args
		want    *mydynamo.Session
		wantErr bool
	}{
		"case 1": {
			args: args{
				ID:        "1234",
				sessionID: "abcde",
				now:       "2025-02-01 12:40:00",
			},
			want:    &session,
			wantErr: false,
		},
		"case 2": {
			args: args{
				ID:        "123",
				sessionID: "abcde",
				now:       "2025-02-01 12:40:00",
			},
			want:    nil,
			wantErr: true,
		},
		"case 3": {
			args: args{
				ID:        "1234",
				sessionID: "abcd",
				now:       "2025-02-01 12:40:00",
			},
			want:    nil,
			wantErr: true,
		},
		"case 4": {
			args: args{
				ID:        "1234",
				sessionID: "abcde",
				now:       "2025-02-01 13:40:00",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for testName, tt := range cases {
		t.Run(testName, func(t *testing.T) {
			got, err := dynamo.GetSession(tt.args.ID, tt.args.sessionID, tt.args.now)
			if (err != nil) != tt.wantErr {
				t.Errorf("DynamoDB.GetSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DynamoDB.GetSession() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDynamoDB_UpdateSession(t *testing.T) {
	ctx := context.Background()

	session := mydynamo.Session{
		ID:        "1234",
		SessionID: "abcde",
		CreatedAt: "2025-02-01 12:34:56",
		Expire:    "2025-02-01 13:34:56",
		TTL:       1738384496,
	}

	dynamo := setupDynamoTest(ctx, t, "Session_UpdateSession")
	saveTestItem(t, dynamo, session)

	now := time.Date(2025, 2, 1, 13, 0, 0, 0, time.Local)
	expectedExpire := now.Add(1 * time.Hour)

	type args struct {
		id        string
		sessionID string
		now       time.Time
	}
	cases := map[string]struct {
		args    args
		want    *mydynamo.Session
		wantErr bool
	}{
		"case 1": {
			args: args{
				id:        "1234",
				sessionID: "abcde",
				now:       now,
			},
			want: &mydynamo.Session{
				ID:        "1234",
				SessionID: "abcde",
				CreatedAt: "2025-02-01 12:34:56",
				Expire:    expectedExpire.Format(time.DateTime),
				TTL:       int(expectedExpire.Unix()),
			},
			wantErr: false,
		},
		"case 2": {
			args: args{
				id:        "123",
				sessionID: "abcde",
				now:       now,
			},
			want:    nil,
			wantErr: true,
		},
		"case 3": {
			args: args{
				id:        "1234",
				sessionID: "abcd",
				now:       now,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for testName, tt := range cases {
		t.Run(testName, func(t *testing.T) {
			got, err := dynamo.UpdateSession(tt.args.id, tt.args.sessionID, tt.args.now)
			if (err != nil) != tt.wantErr {
				t.Errorf("DynamoDB.UpdateSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func SetupLocalStack(ctx context.Context, t *testing.T) (string, func()) {
	t.Helper()
	// testcontainers で LocalStack を起動
	c, err := localstack.Run(ctx, "localstack/localstack:latest")
	if err != nil {
		t.Fatalf("Failed to start LocalStack container: %v", err)
	}

	// エンドポイント取得
	awsEndpoint, err := c.PortEndpoint(ctx, "4566", "http")
	if err != nil {
		t.Fatalf("Failed to get LocalStack endpoint: %v", err)
	}

	t.Logf("Started LocalStack at %s", awsEndpoint)

	// クリーンアップ関数を設定
	cleanup := func() {
		_ = c.Terminate(ctx)
	}

	return awsEndpoint, cleanup
}

func setupDynamoTest(ctx context.Context, t *testing.T, tableName string) *mydynamo.DynamoDB {
	t.Helper()

	// DynamoDBクライアントの初期化
	dynamo, err := mydynamo.NewDynamoDB(endpoint, tableName)
	if err != nil {
		t.Logf("Failed to init DynamoDB client: %v\n", err)
	}

	// テスト用テーブルの作成
	_, err = dynamo.Client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName:   aws.String(tableName),
		BillingMode: types.BillingModePayPerRequest,
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("sessionId"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("sessionId"),
				KeyType:       types.KeyTypeRange,
			},
		},
	})

	if err != nil {
		t.Logf("Failed to create test table: %v\n", err)
	}

	return dynamo
}

func saveTestItem(t *testing.T, dynamo *mydynamo.DynamoDB, item any) {
	value, err := attributevalue.MarshalMap(item)
	if err != nil {
		t.Logf("marshal value error: %v", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: &dynamo.Table,
		Item:      value,
	}

	if _, err := dynamo.Client.PutItem(context.TODO(), input); err != nil {
		t.Logf("put item error: %v", err)
	}
}
