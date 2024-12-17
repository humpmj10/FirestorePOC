package main

import (
	"context"
	"github.com/google/uuid"
	"log"
	"time"

	"cloud.google.com/go/firestore"

	"FirestorePOC/DAO"
	"FirestorePOC/models/transaction"
)

func main() {
	ctx := context.Background()

	client, err := firestore.NewClient(ctx, "firestore-poc-444422")
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer client.Close()

	// Initialize the DAO
	transactionDAO := DAO.NewFirestoreDAO(client)

	transIDs := buildTransactionIDs(3)

	// Upsert a single transaction
	log.Println("Upserting a single transaction...")
	transactionID := transIDs[0]
	transaction1 := transaction.Model{
		ID:            transactionID,
		AcctID:        "acct_001",
		PostedTime:    time.Now(),
		CardNumber:    "4111111111111111",
		Type:          "PURCHASE",
		OnlineService: []string{"Amazon Web Services"},
	}
	if err := transactionDAO.UpsertTransaction(ctx, transactionID, transaction1); err != nil {
		log.Fatalf("Error upserting transaction: %v", err)
	}
	log.Println("Transaction upserted:", transaction1)

	transaction1.AcctID = "acct_010"
	transaction1.Type = "REFUND"
	if err := transactionDAO.UpsertTransactionWithHistory(ctx, transactionID, transaction1); err != nil {
		log.Fatalf("Error upserting transaction with history: %v", err)
	}
	log.Println("Transaction upserted with history:", transaction1)

	// Upsert multiple transactions
	log.Println("Upserting multiple transactions...")
	transactions := []transaction.Model{
		{
			ID:            transIDs[1],
			AcctID:        "acct_002",
			PostedTime:    time.Now(),
			CardNumber:    "4222222222222222",
			Type:          "REFUND",
			OnlineService: []string{"Google Firestore", "Apple"},
		},
		{
			ID:         transIDs[2],
			AcctID:     "acct_003",
			PostedTime: time.Now(),
			CardNumber: "4333333333333333",
			Type:       "PURCHASE",
		},
	}
	transactionIDs := []string{transIDs[1], transIDs[2]}
	if err := transactionDAO.UpsertTransactions(ctx, transactionIDs, transactions); err != nil {
		log.Fatalf("Error upserting multiple transactions: %v", err)
	}
	log.Println("Multiple transactions upserted:", transactions)

	// Fetch a single transaction
	log.Println("Fetching a single transaction...")
	fetchedTransaction, err := transactionDAO.GetTransaction(ctx, transactionID)
	if err != nil {
		log.Fatalf("Error fetching transaction: %v", err)
	}
	log.Println("Fetched transaction:", fetchedTransaction)

	// Fetch a single transaction
	log.Println("Fetching a single transaction...")
	fetchedTransactionWithFields, err := transactionDAO.GetTransactionWithFields(ctx, transactionID, []string{"AcctID", "PostedTime"})
	if err != nil {
		log.Fatalf("Error fetching transaction: %v", err)
	}
	log.Println("Fetched transaction:", fetchedTransactionWithFields)

	// Fetch multiple transactions
	log.Println("Fetching multiple transactions...")
	fetchedTransactions, err := transactionDAO.GetTransactions(ctx, transactionIDs)
	if err != nil {
		log.Fatalf("Error fetching transactions: %v", err)
	}
	log.Println("Fetched transactions:", fetchedTransactions)

	// Search transactions with a query
	log.Println("Searching transactions...")
	query := transaction.Query{
		//AcctId: "acct_001",
		//Types:  []string{"PURCHASE"},
		//StartTime: time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
		//EndTime:   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		//Limit:            10,
		OnlineService: []string{"Google Firestore"},
	}
	searchResults, err := transactionDAO.SearchTransactions(ctx, query)
	if err != nil {
		log.Fatalf("Error searching transactions: %v", err)
	}
	log.Println("Search results:", searchResults)
}

func buildTransactionIDs(num int) []string {
	transactionIDs := make([]string, num)
	for i := 0; i < num; i++ {
		transactionIDs[i] = uuid.NewString()
	}

	return transactionIDs
}
