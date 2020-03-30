package schema

type SymptomType string

const (
	SymptomFever     SymptomType = "fever"
	SymptomCough     SymptomType = "cough"
	SymptomFatigue   SymptomType = "fatigue"
	SymptomBreathing SymptomType = "breathing"
	SymptomNasal     SymptomType = "nasal"
	SymptomThroat    SymptomType = "throat"
	SymptomChest     SymptomType = "chest"
	SymptomFace      SymptomType = "face"
)

type Symptom struct {
	ID   SymptomType `json:"id"`
	Name string      `json:"name"`
	Desc string      `json:"desc"`
}

var Symptoms = []Symptom{
	{SymptomFever, "Fever", ""},
	{SymptomCough, "Cough", ""},
	{SymptomFatigue, "Fatigue", ""},
	{SymptomBreathing, "Difficulty breathing", ""},
	{SymptomNasal, "Nasal congestion", ""},
	{SymptomThroat, "Sore throat", ""},
	{SymptomChest, "Chest pain or pressure", ""},
	{SymptomFace, "Bluish lips or face", ""},
}
