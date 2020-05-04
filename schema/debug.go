package schema

type Debug struct {
	Metrics  Metric `json:"metrics"`
	Users    int    `json:"users"`
	AQI      int    `json:"aqi"`
	Symptoms int    `json:"symptoms"`
}
