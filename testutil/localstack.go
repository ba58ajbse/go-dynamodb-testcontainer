package testutil

import (
	"context"
	"log"

	"github.com/floci-io/testcontainers-floci-go/flociaws"
)

func SetupLocalStack(ctx context.Context) (string, func()) {
	// testcontainers で Floci (LocalStack 互換の AWS エミュレータ) を起動
	c, err := flociaws.Run(ctx, "floci/floci:2.3.0", flociaws.WithRegion("ap-northeast-1"))
	if err != nil {
		log.Fatalf("Failed to start Floci container: %v", err)
	}

	// エンドポイント取得
	awsEndpoint := c.GetEndpoint()

	log.Printf("Started Floci at %s", awsEndpoint)

	// クリーンアップ関数を設定
	cleanup := func() {
		_ = c.Terminate(ctx)
	}

	return awsEndpoint, cleanup
}
