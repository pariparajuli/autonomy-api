package schema

import (
	"encoding/json"
)

type SymptomType string

// SymptomFromID is a map which key is Symptom.ID and value is a object of Symptom
var SymptomFromID = map[SymptomType]Symptom{
	SymptomType(Fever):   COVID19Symptoms[0],
	SymptomType(Cough):   COVID19Symptoms[1],
	SymptomType(Fatigue): COVID19Symptoms[2],
	SymptomType(Breath):  COVID19Symptoms[3],
	SymptomType(Nasal):   COVID19Symptoms[4],
	SymptomType(Throat):  COVID19Symptoms[5],
	SymptomType(Chest):   COVID19Symptoms[6],
	SymptomType(Face):    COVID19Symptoms[7],
}

type SymptomSource string

const (
	OfficialSymptom   SymptomSource = "official"
	SuggestedSymptom  SymptomSource = "suggested"
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
	Weight float64       `json:"-" bson:"weight"`
}

// The system defined symptoms. The list will be inserted into database by migration function
var COVID19Symptoms = []Symptom{
	{Fever, "Fever", "Body temperature above 100ºF (38ºC)", OfficialSymptom, 2},
	{Cough, "Dry cough", "Without mucous or phlegm (rattling)", OfficialSymptom, 2},
	{Fatigue, "Fatigue or tiredness", "Unusual lack of energy or feeling run down", OfficialSymptom, 1},
	{Breath, "Shortness of breath", "Constriction or difficulty inhaling fully", OfficialSymptom, 1},
	{Nasal, "Nasal congestion", "Stuffy or blocked nose", OfficialSymptom, 1},
	{Throat, "Sore throat", "Throat pain, scratchiness, or irritation", OfficialSymptom, 1},
	{Chest, "Chest pain", "Persistent pain or pressure in the chest", OfficialSymptom, 1},
	{Face, "Bluish lips or face", "Not caused by cold exposure", OfficialSymptom, 1},
}

var GeneralSymptoms = []Symptom{
	{ID: "insomnia", Name: "Insomnia", Source: SuggestedSymptom, Weight: 1},
}

// SymptomReportData the struct to store symptom data and score
type SymptomReportData struct {
	ProfileID          string    `json:"profile_id" bson:"profile_id"`
	AccountNumber      string    `json:"account_number" bson:"account_number"`
	OfficialSymptoms   []Symptom `json:"official_symptoms" bson:"official_symptoms"`
	CustomizedSymptoms []Symptom `json:"customized_symptoms" bson:"customized_symptoms"`
	Location           GeoJSON   `json:"location" bson:"location"`
	Timestamp          int64     `json:"ts" bson:"ts"`
	SymptomScore       float64   `json:"score" bson:"score"`
}

type SymptomDistribution map[SymptomType]int

func (s *SymptomReportData) MarshalJSON() ([]byte, error) {
	allSymptoms := append(s.OfficialSymptoms, s.CustomizedSymptoms...)
	if allSymptoms == nil {
		allSymptoms = make([]Symptom, 0)
	}
	return json.Marshal(&struct {
		Symptoms  []Symptom `json:"symptoms"`
		Location  Location  `json:"location"`
		Timestamp int64     `json:"timestamp"`
	}{
		Symptoms:  allSymptoms,
		Location:  Location{Longitude: s.Location.Coordinates[0], Latitude: s.Location.Coordinates[1]},
		Timestamp: s.Timestamp,
	})
}
