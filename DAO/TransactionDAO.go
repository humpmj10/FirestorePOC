package DAO

import (
	"FirestorePOC/models/comparison"
	"FirestorePOC/models/transaction"
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	"fmt"
	"google.golang.org/api/iterator"
	"log"
	"slices"
	"time"
)

const transactionCollection = "Transactions"

type DAO interface {
	GetTransaction(ctx context.Context, id string) (transaction.Model, error)
	GetTransactionWithFields(ctx context.Context, id string, fields []string) (transaction.Model, error)
	GetTransactions(ctx context.Context, ids []string) ([]transaction.Model, error)
	UpsertTransaction(ctx context.Context, id string, model transaction.Model) error
	UpsertTransactionWithHistory(ctx context.Context, id string, model transaction.Model) error
	UpsertTransactions(ctx context.Context, ids []string, models []transaction.Model) error
	SearchTransactions(ctx context.Context, q transaction.Query) ([]transaction.Model, error)
}

type FirestoreDAO struct {
	client *firestore.Client
}

func NewFirestoreDAO(client *firestore.Client) FirestoreDAO {
	return FirestoreDAO{client: client}
}

func (d *FirestoreDAO) GetTransaction(ctx context.Context, id string) (transaction.Model, error) {
	docRef := d.client.Collection(transactionCollection).Doc(id)
	doc, err := docRef.Get(ctx)
	if err != nil {
		return transaction.Model{}, err
	}

	var model transaction.Model
	if err := doc.DataTo(&model); err != nil {
		return transaction.Model{}, err
	}

	return model, nil
}

var transactionFields = []string{"AcctID", "CardNumber", "OnlineService", "PostedTime", "Type"}

// GetTransactionWithFields Example on how to only retrieve certain fields, it doesn't help performance as firestore ways grabs the entire document
// however, we can support only certain fields being return in the response
func (d *FirestoreDAO) GetTransactionWithFields(ctx context.Context, id string, fields []string) (transaction.Model, error) {
	// Ensure we have fields to pass to firestore, else just call get transaction normally
	if len(fields) == 0 {
		return d.GetTransaction(ctx, id)
	}

	// Validate the fields
	for _, field := range fields {
		if !slices.Contains(transactionFields, field) {
			return transaction.Model{}, fmt.Errorf("field %s not found in accectable fields", field)
		}
	}

	// Only select the fields passed in
	query := d.client.Collection(transactionCollection).Where("ID", comparison.OpEqual, id).Select(fields...)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var model transaction.Model
	doc, err := iter.Next()
	if errors.Is(err, iterator.Done) {
		return model, fmt.Errorf(`item not found: "%s" not found`, id)
	}
	if err != nil {
		return model, err
	}

	if err := doc.DataTo(&model); err != nil {
		return model, err
	}

	return model, nil
}

func (d *FirestoreDAO) GetTransactions(ctx context.Context, ids []string) ([]transaction.Model, error) {

	var docRefs []*firestore.DocumentRef
	for _, id := range ids {
		docRefs = append(docRefs, d.client.Collection(transactionCollection).Doc(id))
	}

	// Perform a batched read
	docs, err := d.client.GetAll(ctx, docRefs)
	if err != nil {
		return nil, err
	}

	var results []transaction.Model
	for _, doc := range docs {
		var model transaction.Model
		if err := doc.DataTo(&model); err != nil {
			return nil, err
		}
		results = append(results, model)
	}

	return results, nil
}

func (d *FirestoreDAO) UpsertTransaction(ctx context.Context, id string, model transaction.Model) error {
	_, err := d.client.Collection(transactionCollection).Doc(id).Set(ctx, model)
	return err
}

func (d *FirestoreDAO) UpsertTransactionWithHistory(ctx context.Context, id string, model transaction.Model) error {
	mainDocRef := d.client.Collection(transactionCollection).Doc(id)
	historyCollRef := mainDocRef.Collection("history")

	// Using transaction to ensure that the main document and appending to version history happen atomically
	err := d.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// Step 1: fetch the current version
		docSnap, err := tx.Get(mainDocRef)
		if err != nil {
			return fmt.Errorf("failed to get main document %v", err)
		}

		currentVersion := 0
		if docSnap.Exists() {
			// Get the current version of the document
			if v, ok := docSnap.Data()["version"].(int64); ok {
				currentVersion = int(v)
			}
		}

		// Step 2: Increment the version
		newVersion := currentVersion + 1
		model.Version = currentVersion
		model.LastUpdated = time.Now()

		// Step 3: Update the main document with the latest data
		err = tx.Set(mainDocRef, model)
		if err != nil {
			return fmt.Errorf("failed to set main document %v", err)
		}

		// Step 4: Append to the history sub-collection
		historyDocRef := historyCollRef.Doc(fmt.Sprintf("version_%d", newVersion))
		historyData := map[string]interface{}{
			accountIDField:     model.AcctID,
			CardNumberField:    model.CardNumber,
			OnlineServiceField: model.OnlineService,
			postedTimeField:    model.PostedTime,
			typeField:          model.Type,
		}
		err = tx.Set(historyDocRef, historyData)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to upsert transaction with history: %w", err)
	}

	log.Printf("Successfully upserted document %s with version %d\n", id, model.Version)
	return nil
}

func (d *FirestoreDAO) UpsertTransactions(ctx context.Context, ids []string, models []transaction.Model) error {
	if len(ids) != len(models) {
		return errors.New("the number of ids does not match the number of documents")
	}

	// Create a BulkWriter Instance, it doesn't natively support automatic flushing after time/size thresholds are met
	writer := d.client.BulkWriter(ctx)

	// Add operations to BulkWriter
	for i, id := range ids {
		docRef := d.client.Collection(transactionCollection).Doc(id)

		// Perform Set operation
		_, err := writer.Set(docRef, models[i])
		if err != nil {
			return err
		}
	}

	// It's up to use to flush once we are finished adding documents to the buffer
	writer.Flush()

	log.Println("All transactions upserted successfully!")
	return nil
}

// Fields are case-sensitive in firestore, adding these constants to avoid typos
const (
	accountIDField     = "AcctID"
	typeField          = "Type"
	postedTimeField    = "PostedTime"
	onlineServiceField = "OnlineService"
	VersionField       = "Version"
	LastUpdatedField   = "LastUpdated"
	CardNumberField    = "CardNumber"
	OnlineServiceField = "OnlineService"
)

func (d *FirestoreDAO) SearchTransactions(ctx context.Context, q transaction.Query) ([]transaction.Model, error) {
	collectionRef := d.client.Collection(transactionCollection)
	query := collectionRef.Query

	if q.AcctId != "" {
		query = query.Where(accountIDField, comparison.OpEqual, q.AcctId)
	}

	if len(q.Types) > 0 {
		query = query.Where(typeField, comparison.OpIn, q.Types)
	}

	if q.StartTime != "" && q.EndTime != "" {
		startTime, err := time.Parse(time.RFC3339, q.StartTime)
		if err != nil {
			return nil, fmt.Errorf("invalid start time: %w", err)
		}
		endTime, err := time.Parse(time.RFC3339, q.EndTime)
		if err != nil {
			return nil, fmt.Errorf("invalid end time: %w", err)
		}
		query = query.Where(postedTimeField, comparison.OpGreaterThanOrEqual, startTime).Where(postedTimeField, comparison.OpLessThan, endTime)
	}

	// Add limit and offset
	if q.Limit > 0 {
		query = query.Limit(q.Limit)
	}

	if len(q.OnlineService) > 0 {
		// Use array-contains-any and not array-contains, because we are passing an array of search values
		query = query.Where(onlineServiceField, comparison.OpArrayContainsAny, q.OnlineService)
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	var results []transaction.Model
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate documents: %w", err)
		}

		var model transaction.Model
		if err := doc.DataTo(&model); err != nil {
			return nil, fmt.Errorf("failed to map document data: %w", err)
		}
		results = append(results, model)
	}

	return results, nil
}
