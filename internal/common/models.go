package common

type Operation struct {
	ID            int `gorm:"primaryKey"`
	DeviceToken   string
	GroupId       string
	OperationType string
	Sql           string
	Args          string
	CreatedAt     int64
}

type RelatedEntity struct {
	OperationID int
	EntityID    string
	EntityName  string
}
