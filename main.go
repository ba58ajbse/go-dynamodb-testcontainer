package main

import (
	"fmt"
	mydynamo "godynamodb/dynamodb"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

/*
DYNAMODB_ENDPOINT=http://localhost:4566 DYNAMODB_TABLE=Session go run .
*/
func main() {
	table := os.Getenv("DYNAMODB_TABLE")

	// DynamoDB クライアントを作成
	client, err := mydynamo.NewDynamoDB(table)
	if err != nil {
		log.Fatalf("Failed to init DynamoDB client: %v", err)
	}

	sessionID := strconv.Itoa(rand.New(rand.NewSource(time.Now().UnixNano())).Int())
	// sessionID := "abcde"
	session := mydynamo.NewSession("1234", sessionID)

	// create
	if err := client.SaveSession(session); err != nil {
		log.Fatalf("Failed to create item: %v", err)
	}

	time.Sleep(5 * time.Second)

	// get
	newSession, err := client.GetSession("1234", sessionID, time.Now().Local().Format(time.DateTime))
	if err != nil {
		log.Fatalf("Failed to get item: %v", err)
	}

	fmt.Printf("get session: %+v\n", newSession)

	nowTime := time.Now().Local()
	// update
	ret, err := client.UpdateSession("1234", sessionID, nowTime)
	if err != nil {
		log.Fatalf("Failed to update item: %v", err)
	}

	fmt.Println("Updated:", &ret)
}
