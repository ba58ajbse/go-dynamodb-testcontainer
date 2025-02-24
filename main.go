package main

import (
	"fmt"
	mydynamo "godynamodb/dynamodb"
	"log"
	"os"
	"time"
)

/*
DYNAMODB_ENDPOINT=http://localhost:4566 DYNAMODB_TABLE=Session go run .
*/
func main() {
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	table := os.Getenv("DYNAMODB_TABLE")

	// DynamoDB クライアントを作成
	client, err := mydynamo.NewDynamoDB(endpoint, table)
	if err != nil {
		log.Fatalf("Failed to init DynamoDB client: %v", err)
	}

	session := mydynamo.NewSession("1234", "abcde")

	// create
	if err := client.SaveSession(session); err != nil {
		log.Fatalf("Failed to create item: %v", err)
	}

	time.Sleep(5 * time.Second)

	// get
	newSession, err := client.GetSession("1234", "abcde", time.Now().Local().Format(time.DateTime))
	if err != nil {
		log.Fatalf("Failed to get item: %v", err)
	}

	fmt.Printf("get session: %+v\n", newSession)

	nowTime := time.Now().Local()
	// update
	ret, err := client.UpdateSession("1234", "abcde", nowTime)
	if err != nil {
		log.Fatalf("Failed to update item: %v", err)
	}

	fmt.Println("Updated:", &ret)
}
