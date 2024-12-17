package comparison

// Note there are limits to the number of operators and disjunctions that can be used in one query, see docs for limits
const (
	OpEqual              = "=="
	OpNotEqual           = "!="
	OpLessThan           = "<"
	OpLessThanOrEqual    = "<="
	OpGreaterThan        = ">"
	OpGreaterThanOrEqual = ">="
	OpIn                 = "in"                 // Used if the field can contain any of these values, https://firebase.google.com/docs/firestore/query-data/queries#in_not-in_and_array-contains-any
	OpNotIn              = "not-in"             // Used as exclusion list, if field has this value exclude from hits, check docs https://firebase.google.com/docs/firestore/query-data/queries#not-in
	OpArrayContains      = "array-contains"     // Use if just passing one value to as filter criteria, check the docs for more info https://firebase.google.com/docs/firestore/query-data/queries#array_membership
	OpArrayContainsAny   = "array-contains-any" // Use if passing an array of values as filter criteria, check the docs for more info https://firebase.google.com/docs/firestore/query-data/queries#array-contains-any
)
