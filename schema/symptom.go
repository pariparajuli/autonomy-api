package schema

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

const (
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
	ID     SymptomType `json:"id"`
	Name   string      `json:"name"`
	Desc   string      `json:"desc"`
	Weight float64     `json:"-"`
}

var Symptoms = []Symptom{
	{Fever, "Fever", "Body temperature above 100ºF (38ºC)", 2},
	{Cough, "Dry cough", "Without mucous or phlegm (rattling)", 1},
	{Fatigue, "Fatigue or tiredness", "Unusual lack of energy or feeling run down", 1},
	{Breath, "Shortness of breath", "Constriction or difficulty inhaling fully", 1},
	{Nasal, "Nasal congestion", "Stuffy or blocked nose", 1},
	{Throat, "Sore throat", "Throat pain, scratchiness, or irritation", 1},
	{Chest, "Chest pain", "Persistent pain or pressure in the chest", 1},
	{Face, "Bluish lips or face", "Not caused by cold exposure", 1},
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
