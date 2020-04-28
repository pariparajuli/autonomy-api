package schema

import (
	"encoding/json"
)

type SymptomType string

// SymptomFromID is a map which key is Symptom.ID and value is a object of Symptom
var SymptomFromID = map[SymptomType]Symptom{
	SymptomType(Fever):   Symptoms[0],
	SymptomType(Cough):   Symptoms[1],
	SymptomType(Fatigue): Symptoms[2],
	SymptomType(Breath):  Symptoms[3],
	SymptomType(Nasal):   Symptoms[4],
	SymptomType(Throat):  Symptoms[5],
	SymptomType(Chest):   Symptoms[6],
	SymptomType(Face):    Symptoms[7],
}

type SymptomSource string

const (
	OfficialSymptom   SymptomSource = "official"
	CustomizedSymptom SymptomSource = "customized"
)

const (
	SymptomCollection       = "symptom"
	SymptomReportCollection = "symptomReport"
	TotalSymptomWeight      = 9
)

const (
	Fever   SymptomType = "fever"
	Cough   SymptomType = "cough"
	Fatigue SymptomType = "fatigue"
	Breath  SymptomType = "breath"
	Nasal   SymptomType = "nasal"
	Throat  SymptomType = "throat"
	Chest   SymptomType = "chest"
	Face    SymptomType = "face"
)

type Symptom struct {
	ID     SymptomType   `json:"id" bson:"_id"`
	Name   string        `json:"name" bson:"name"`
	Desc   string        `json:"desc" bson:"desc"`
	Source SymptomSource `json:"-" bson:"source"`
	Weight float64       `json:"-" bson:"-"`
}

// The system defined symptoms. The list will be inserted into database by migration function
var Symptoms = []Symptom{
	{Fever, "Fever", "Body temperature above 100ºF (38ºC)", OfficialSymptom, 2},
	{Cough, "Dry cough", "Without mucous or phlegm (rattling)", OfficialSymptom, 1},
	{Fatigue, "Fatigue or tiredness", "Unusual lack of energy or feeling run down", OfficialSymptom, 1},
	{Breath, "Shortness of breath", "Constriction or difficulty inhaling fully", OfficialSymptom, 1},
	{Nasal, "Nasal congestion", "Stuffy or blocked nose", OfficialSymptom, 1},
	{Throat, "Sore throat", "Throat pain, scratchiness, or irritation", OfficialSymptom, 1},
	{Chest, "Chest pain", "Persistent pain or pressure in the chest", OfficialSymptom, 1},
	{Face, "Bluish lips or face", "Not caused by cold exposure", OfficialSymptom, 1},
}

// SymptomReportData the struct to store symptom data and score
type SymptomReportData struct {
	ProfileID     string   `json:"profile_id" bson:"profile_id"`
	AccountNumber string   `json:"account_number" bson:"account_number"`
	Symptoms      []string `json:"symptoms" bson:"symptoms"`
	Location      GeoJSON  `json:"location" bson:"location"`
	SymptomScore  float64  `json:"symptom_score" bson:"symptom_score"`
	Timestamp     int64    `json:"ts" bson:"ts"`
}

func (s *SymptomReportData) MarshalJSON() ([]byte, error) {
	symptoms := s.Symptoms
	if s.Symptoms == nil {
		symptoms = make([]string, 0)
	}
	return json.Marshal(&struct {
		Symptoms  []string `json:"symptoms"`
		Location  Location `json:"location"`
		Timestamp int64    `json:"timestamp"`
	}{
		Symptoms:  symptoms,
		Location:  Location{Longitude: s.Location.Coordinates[0], Latitude: s.Location.Coordinates[1]},
		Timestamp: s.Timestamp,
	})
}
