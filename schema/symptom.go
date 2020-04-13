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
	ID     SymptomType `json:"id"`
	Name   string      `json:"name"`
	Desc   string      `json:"desc"`
	Weight float64     `json:"weight"`
}

var Symptoms = []Symptom{
	{SymptomFever, "Fever", "Body temperature above 100ºF (38ºC)", 2},
	{SymptomCough, "Dry cough", "Without mucous or phlegm (rattling)", 1},
	{SymptomFatigue, "Fatigue or tiredness", "Unusual lack of energy or feeling run down", 1},
	{SymptomBreath, "Shortness of breath", "Constriction or difficulty inhaling fully", 1},
	{SymptomNasal, "Nasal congestion", "Stuffy or blocked nose", 1},
	{SymptomThroat, "Sore throat", "Throat pain, scratchiness, or irritation", 1},
	{SymptomChest, "Chest pain", "Persistent pain or pressure in the chest", 1},
	{SymptomFace, "Bluish lips or face", "Not caused by cold exposure", 1},
}
