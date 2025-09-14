package domain

// FeedType represents different types of feed available
type FeedType struct {
	FeedTypeID      int64
	Name            string
	Description     *string
	NutritionalInfo *string
	Audit           AuditFields
}
