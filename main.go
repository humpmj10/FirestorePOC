package main

import (
	"context"
	"errors"
	"google.golang.org/api/iterator"
	"log"
	"time"

	"cloud.google.com/go/firestore"
)

func main() {
	ctx := context.Background()

	// Create Firestore client
	client, err := firestore.NewClient(ctx, "firestore-poc-444422")
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer client.Close()

	// Test CRUD operations
	docID := "cache_12345"

	if err := createDocument(ctx, client, docID); err != nil {
		log.Fatalf("Create failed: %v", err)
	}

	if err := readDocument(ctx, client, docID); err != nil {
		log.Fatalf("Read failed: %v", err)
	}

	if err := updateDocument(ctx, client, docID); err != nil {
		log.Fatalf("Update failed: %v", err)
	}

	//if err := deleteDocument(ctx, client, docID); err != nil {
	//	log.Fatalf("Delete failed: %v", err)
	//}

	if err := fetchAllDocuments(ctx, client); err != nil {
		log.Fatalf("Fetch All failed: %v", err)
	}

	if err := filterDocuments(ctx, client); err != nil {
		log.Fatalf("Filter failed: %v", err)
	}

	if err := queryWithRange(ctx, client); err != nil {
		log.Fatalf("Range Query failed: %v", err)
	}
}

func createDocument(ctx context.Context, client *firestore.Client, docID string) error {
	_, err := client.Collection("apiCache").Doc(docID).Set(ctx, map[string]interface{}{
		"url":       "https://example.com/api/resource",
		"response":  "{\"status\":200,\"data\":{\"key\":\"value\"}}",
		"timestamp": firestore.ServerTimestamp,
	})
	if err != nil {
		return err
	}
	log.Printf("Document %s created successfully!", docID)
	return nil
}

func readDocument(ctx context.Context, client *firestore.Client, docID string) error {
	doc, err := client.Collection("apiCache").Doc(docID).Get(ctx)
	if err != nil {
		return err
	}
	log.Printf("Document Data: %v", doc.Data())
	return nil
}

func updateDocument(ctx context.Context, client *firestore.Client, docID string) error {
	_, err := client.Collection("apiCache").Doc(docID).Update(ctx, []firestore.Update{
		{Path: "response", Value: "{\"status\":200,\"data\":{\"key\":\"updated\"}}"},
		{Path: "timestamp", Value: firestore.ServerTimestamp},
	})
	if err != nil {
		return err
	}
	log.Printf("Document %s updated successfully!", docID)
	return nil
}

func deleteDocument(ctx context.Context, client *firestore.Client, docID string) error {
	_, err := client.Collection("apiCache").Doc(docID).Delete(ctx)
	if err != nil {
		return err
	}
	log.Printf("Document %s deleted successfully!", docID)
	return nil
}

func fetchAllDocuments(ctx context.Context, client *firestore.Client) error {
	iter := client.Collection("apiCache").Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return err
		}
		log.Printf("Document ID: %s, Data: %v", doc.Ref.ID, doc.Data())
	}
	return nil
}

func filterDocuments(ctx context.Context, client *firestore.Client) error {
	iter := client.Collection("apiCache").Where("status", "==", 200).Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return err
		}
		log.Printf("Filtered Document ID: %s, Data: %v", doc.Ref.ID, doc.Data())
	}
	return nil
}

func queryWithRange(ctx context.Context, client *firestore.Client) error {
	// Define a concrete start time for the range
	startTime := time.Now().Add(-24 * time.Hour) // 24 hours ago

	iter := client.Collection("apiCache").
		Where("timestamp", ">=", startTime).
		Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return err
		}
		log.Printf("Range Query Document ID: %s, Data: %v", doc.Ref.ID, doc.Data())
	}
	return nil
}
