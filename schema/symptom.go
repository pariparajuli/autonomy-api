package schema

type SymptomType string

const (
	SymptomFever   SymptomType = "fever"
	SymptomCough   SymptomType = "cough"
	SymptomFatigue SymptomType = "fatigue"
	SymptomBreath  SymptomType = "breath"
	SymptomNasal   SymptomType = "nasal"
	SymptomThroat  SymptomType = "throat"
	SymptomChest   SymptomType = "chest"
	SymptomFace    SymptomType = "face"
)

type Symptom struct {
	ID   SymptomType `json:"id"`
	Name string      `json:"name"`
	Desc string      `json:"desc"`
}

var Symptoms = []Symptom{
	{SymptomFever, "Fever", "Body temperature above 100ºF (38ºC)"},
	{SymptomCough, "Dry cough", "Without mucous or phlegm (rattling)"},
	{SymptomFatigue, "Fatigue or tiredness", "Unusual lack of energy or feeling run down"},
	{SymptomBreath, "Shortness of breath", "Constriction or difficulty inhaling fully"},
	{SymptomNasal, "Nasal congestion", "Stuffy or blocked nose"},
	{SymptomThroat, "Sore throat", "Throat pain, scratchiness, or irritation"},
	{SymptomChest, "Chest pain", "Persistent pain or pressure in the chest"},
	{SymptomFace, "Bluish lips or face", "Not caused by cold exposure"},
}
