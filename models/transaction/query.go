package transaction

type Query struct {
	AcctId        string
	Types         []string
	OnlineService []string
	StartTime     string
	EndTime       string
	Limit         int
	Offset        int // not implemented, currently Firestore doesn't support offset natively
}
