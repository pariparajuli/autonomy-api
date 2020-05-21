package schema

import (
	"encoding/json"
)

type SymptomType string

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

var (
	OfficialSymptoms = map[string]bool{
		"fever":   true,
		"cough":   true,
		"fatigue": true,
		"breath":  true,
		"nasal":   true,
		"throat":  true,
		"chest":   true,
		"face":    true,
	}
)

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
	{ID: "suggestion_1", Name: "Abdominal bloating", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_2", Name: "Abdominal pain (stomachache)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_3", Name: "Abnormal amount of body hair growth (hypertrichosis)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_4", Name: "Abnormal skin tingling, prickling, chilling, burning, or numbness (paresthesia)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_5", Name: "Abnormal sweating (hyperhidrosis)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_6", Name: "Abnormal sweating (perspiration)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_7", Name: "Abnormal vaginal bleeding", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_8", Name: "Abnormal walking (ataxia)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_9", Name: "Abnormally fast breathing (tachypnea)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_10", Name: "Abnormally slow breathing (bradypnea)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_11", Name: "Absence of menstrual period (amenorrhea)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_12", Name: "Anxiety", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_13", Name: "Apathy", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_14", Name: "Back pain (backache)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_15", Name: "Bad breath (halitosis)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_16", Name: "Blister", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_17", Name: "Blood in semen (hematospermia)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_18", Name: "Blood in stool", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_19", Name: "Blood in urine (hematuria)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_20", Name: "Blurred vision", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_21", Name: "Bruising (contusions, hematoma)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_22", Name: "Burping (belching)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_23", Name: "Chills", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_24", Name: "Chronic pain", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_25", Name: "Confusion", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_26", Name: "Convulsions", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_27", Name: "Coughed up mucous (sputum)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_28", Name: "Coughing up blood (hemoptysis)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_29", Name: "Decreased appetite (anorexia)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_30", Name: "Deformity", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_31", Name: "Delusions", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_32", Name: "Depression", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_33", Name: "Diarrhea", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_34", Name: "Difficult or painful urination (dysuria)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_35", Name: "Difficulty swallowing (dysphagia)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_36", Name: "Dilated pupils", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_37", Name: "Discharge of mucous or pus", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_38", Name: "Dizziness (vertigo)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_39", Name: "Double vision", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_40", Name: "Dry mouth (xerostomia)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_41", Name: "Ear pain (Earache, otalgia)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_42", Name: "Erectile dysfunction (impotence)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_43", Name: "Erratic heartbeat (heart palpitations)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_44", Name: "Excessive body hair (hirsutism)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_45", Name: "Excessive urination (polyuria)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_46", Name: "Eyelid spasms", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_47", Name: "Fainting (loss of consciousness, syncope)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_48", Name: "Feeling of incomplete defecation (rectal tenesmus)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_49", Name: "Fingernail / toenail infection or deformity", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_50", Name: "Flatulence (gas)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_51", Name: "Frequent urination", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_52", Name: "Gastrointestinal bleeding (hemorrhaging)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_53", Name: "Hair loss (alopecia)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_54", Name: "Hallucination", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_55", Name: "Headache", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_56", Name: "Hearing loss", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_57", Name: "Heart arrhythmia (cardiac arrhythmia)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_58", Name: "Heartburn (pyrosis)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_59", Name: "High resting heart rate (tachycardia)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_60", Name: "Inability or weakness moving one side of the body", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_61", Name: "Inability to uninate normally (urinary retention)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_62", Name: "Indigestion (dyspepsia)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_63", Name: "Infertility", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_64", Name: "Involuntary body movements", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_65", Name: "Involuntary eye movements", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_66", Name: "Involuntary urination (urinary incontinence)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_67", Name: "Itching", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_68", Name: "Joint pain (arthralgia)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_69", Name: "Lightheadedness", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_70", Name: "Loss of bowel control (fecal incontinence)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_71", Name: "Loss of normal bowel movements (constipation)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_72", Name: "Loss of normal speech (aphasia)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_73", Name: "Loss of sense of smell (anosmia)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_74", Name: "Loss of sense of taste (dysgeusia / parageusia)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_75", Name: "Loss of vision (blindness)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_76", Name: "Loss or writing ability (dysgraphia)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_77", Name: "Low body temperature (hypothermia)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_78", Name: "Low resting heart rate (bradycardia)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_79", Name: "Malaise", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_80", Name: "Memory loss (amnesia)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_81", Name: "Muscle cramps", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_82", Name: "Muscle loss (cachexia)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_83", Name: "Nausea", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_84", Name: "Nervous tics", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_85", Name: "Nosebleed (epistaxis)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_86", Name: "Oily or fatty feces (steatorrhea)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_87", Name: "Pain when swallowing (odynophagia)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_88", Name: "Painful intercourse", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_89", Name: "Pelvic pain", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_90", Name: "Rash", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_91", Name: "Rectal or anal pain (proctologia fugax)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_92", Name: "Reduced ability to experience pleasure", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_93", Name: "Reduced ability to open the jaws (lockjaw, trismus)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_94", Name: "Ringing or hissing in ears (tinnitus)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_95", Name: "Runny nose (rhinorrhea)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_96", Name: "Sciatica (pain going down the leg from lower back)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_97", Name: "Sharp chest pains while breathing (pleuritic chest pain)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_98", Name: "Shivering", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_99", Name: "Skin pain (dermatome)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_100", Name: "Sleepiness (somnolence)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_101", Name: "Sleeplessness (insomnia)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_102", Name: "Sounds are too loud", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_103", Name: "Stopping of breathing (apnea)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_104", Name: "Stopping of breathing while sleeping (sleep apnea)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_105", Name: "Swelling (edema)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_106", Name: "Swollen or painful lymph nodes (lymphadenopathy / adenopathy)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_107", Name: "Thirst", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_108", Name: "Thoughts of killing yourself (suicidal ideation)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_109", Name: "Tooth pain (toothache)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_110", Name: "Tremors", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_111", Name: "Urethral discharge", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_112", Name: "Vaginal discharge", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_113", Name: "Vomiting", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_114", Name: "Vomiting blood (hematemesis)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_115", Name: "Weakness (muscle weakness)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_116", Name: "Weight gain", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_117", Name: "Weight loss", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_118", Name: "Wound (laceration)", Source: SuggestedSymptom, Weight: 1},
	{ID: "suggestion_119", Name: "Yellowish or greenish skin (jaundice)", Source: SuggestedSymptom, Weight: 1},
}

// SymptomReportData the struct to store symptom data and score
type SymptomReportData struct {
	ProfileID          string    `json:"profile_id" bson:"profile_id"`
	AccountNumber      string    `json:"account_number" bson:"account_number"`
	OfficialSymptoms   []Symptom `json:"official_symptoms" bson:"official_symptoms"`
	CustomizedSymptoms []Symptom `json:"customized_symptoms" bson:"customized_symptoms"`
	Location           GeoJSON   `json:"location" bson:"location"`
	Timestamp          int64     `json:"ts" bson:"ts"`
}

type SymptomDistribution map[string]int

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
