package models

// DashboardCounts holds the count data for the dashboard
type DashboardCounts struct {
	Barns             int64
	FeedTypes         int64
	Staff             int64
	Flocks            int64
	FeedingRecords    int64
	HealthChecks      int64
	MortalityRecords  int64
	ProductionBatches int64
	SlaughterRecords  int64
	InventoryItems    int64
	Customers         int64
	Orders            int64
	OrderItems        int64
}
