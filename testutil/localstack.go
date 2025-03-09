package testutil

import (
	"context"
	"log"

	"github.com/testcontainers/testcontainers-go/modules/localstack"
)

func SetupLocalStack(ctx context.Context) (string, func()) {
	// testcontainers で LocalStack を起動
	c, err := localstack.Run(ctx, "localstack/localstack:latest")
	if err != nil {
		log.Fatalf("Failed to start LocalStack container: %v", err)
	}

	// エンドポイント取得
	awsEndpoint, err := c.PortEndpoint(ctx, "4566", "http")
	if err != nil {
		log.Fatalf("Failed to get LocalStack endpoint: %v", err)
	}

	log.Printf("Started LocalStack at %s", awsEndpoint)

	// クリーンアップ関数を設定
	cleanup := func() {
		_ = c.Terminate(ctx)
	}

	return awsEndpoint, cleanup
}
