#!/bin/bash
# デフォルトリージョンを東京（ap-northeast-1）に設定
export AWS_DEFAULT_REGION=ap-northeast-1

awslocal dynamodb create-table \
    --table-name Session \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=sessionId,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
        AttributeName=sessionId,KeyType=RANGE \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --region ap-northeast-1

awslocal dynamodb update-time-to-live \
    --table-name Session \
    --time-to-live-specification "Enabled=true, AttributeName=ttl" \
    --region ap-northeast-1