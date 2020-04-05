package schema

// CitizenReportData the struct to store citizen data and score
type CitizenReportData struct {
	AccountNumber string    `json:"account_number" bson:"account_number"`
	Symptoms      []Symptom `json:"symptoms" bson:"symptoms"`
	Location      GeoJSON   `json:"location" bson:"location"`
	HealthScore   float64   `json:"health_score" bson:"health_score"`
	Timestamp     int64     `json:"ts" bson:"ts"`
}
