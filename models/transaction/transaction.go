package transaction

import "time"

type Model struct {
	ID            string    `json:"ID,omitempty" firestore:"ID,omitempty"`
	AcctID        string    `json:"account_id,omitempty" firestore:"AcctId,omitempty"`
	PostedTime    time.Time `json:"posted_time,omitempty" firestore:"PostedTime,omitempty"`
	CardNumber    string    `json:"card_number,omitempty" firestore:"CardNumber,omitempty"`
	OnlineService []string  `json:"online_service,omitempty" firestore:"OnlineService,omitempty"`
	Type          string    `json:"type,omitempty" firestore:"Type,omitempty"`
	Version       int       `json:"version,omitempty" firestore:"Version,omitempty"`
	LastUpdated   time.Time `json:"last_updated,omitempty" firestore:"LastUpdated,omitempty"`
}
