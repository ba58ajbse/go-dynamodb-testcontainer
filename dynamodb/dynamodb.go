package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	REGION = "ap-northeast-1"
)

type DynamoDB struct {
	Client *dynamodb.Client
	Table  string
}

type Session struct {
	ID        string `dynamodbav:"id"`
	SessionID string `dynamodbav:"sessionId"`
	CreatedAt string `dynamodbav:"createdAt"`
	Expire    string `dynamodbav:"expire"`
	TTL       int    `dynamodbav:"ttl"`
}

func NewDynamoDB(endpoint, table string) (*DynamoDB, error) {
	// AWS SDK v2 の設定を読み込む
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(REGION), // 必須
		config.WithBaseEndpoint(endpoint),
	)
	if err != nil {
		return nil, err
	}

	if table == "" {
		return nil, errors.New("empty table name")
	}

	// DynamoDB クライアントを作成
	return &DynamoDB{
		Client: dynamodb.NewFromConfig(cfg),
		Table:  table,
	}, nil
}

func NewSession(ID, sessionID string) Session {
	now := time.Now().Local()

	createdAt := now.Format(time.DateTime)
	expire := now.Add(1 * time.Hour)

	return Session{
		ID:        ID,
		SessionID: sessionID,
		CreatedAt: createdAt,
		Expire:    expire.Format(time.DateTime),
		TTL:       int(expire.Unix()),
	}
}

func (d *DynamoDB) SaveSession(s Session) error {
	item, err := attributevalue.MarshalMap(s)
	if err != nil {
		return fmt.Errorf("marshal value error: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: &d.Table,
		Item:      item,
	}

	if _, err := d.Client.PutItem(context.TODO(), input); err != nil {
		return fmt.Errorf("put item error: %w", err)
	}

	return nil
}

func (d *DynamoDB) GetSession(ID, sessionID, now string) (*Session, error) {
	keyCond := expression.KeyAnd(
		expression.Key("id").Equal(expression.Value(ID)),
		expression.Key("sessionId").Equal(expression.Value(sessionID)),
	)
	proj := expression.NamesList(
		expression.Name("id"),
		expression.Name("sessionId"),
		expression.Name("createdAt"),
		expression.Name("expire"),
		expression.Name("ttl"),
	)

	filter := expression.Name("expire").GreaterThan(expression.Value(now))

	builer := expression.NewBuilder().
		WithKeyCondition(keyCond).
		WithProjection(proj).
		WithFilter(filter)
	expr, err := builer.Build()
	if err != nil {
		return nil, fmt.Errorf("build query error: %w", err)
	}

	input := &dynamodb.QueryInput{
		TableName:                 &d.Table,
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		Limit:                     aws.Int32(1),
	}

	res, err := d.Client.Query(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("query response error: %w", err)
	}

	if res.Count == 0 {
		return nil, fmt.Errorf("not found error")
	}
	var session = Session{}
	err = attributevalue.UnmarshalMap(res.Items[0], &session)
	if err != nil {
		return nil, fmt.Errorf("unmarshall list of maps error: %w", err)
	}

	return &session, nil
}

func (d *DynamoDB) UpdateSession(ID, sessionID string, now time.Time) (*Session, error) {
	expire := now.Add(1 * time.Hour)

	update := expression.
		Set(
			expression.Name("expire"),
			expression.Value(expire.Format(time.DateTime)),
		).
		Set(
			expression.Name("ttl"),
			expression.Value(int(expire.Unix())),
		)

	cond := expression.AttributeExists(expression.Name("id")).
		And(expression.AttributeExists(expression.Name("sessionId")))

	expr, err := expression.NewBuilder().
		WithUpdate(update).
		WithCondition(cond).
		Build()
	if err != nil {
		return nil, fmt.Errorf("builed query error: %w", err)
	}

	input := &dynamodb.UpdateItemInput{
		TableName: &d.Table,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{
				Value: ID,
			},
			"sessionId": &types.AttributeValueMemberS{
				Value: sessionID,
			},
		},
		UpdateExpression:          expr.Update(),
		ConditionExpression:       expr.Condition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ReturnValues:              types.ReturnValueAllNew,
	}

	res, err := d.Client.UpdateItem(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	var updatedSession = Session{}
	err = attributevalue.UnmarshalMap(res.Attributes, &updatedSession)
	if err != nil {
		return nil, fmt.Errorf("unmarshall list of maps error: %w", err)
	}

	return &updatedSession, nil
}
