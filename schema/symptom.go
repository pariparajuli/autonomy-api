package schema

import (
	"encoding/json"
)

type ReportType string

const (
	ReportTypeSymptom  = "symptom"
	ReportTypeBehavior = "behavior"
)

type SymptomSource string

const (
	OfficialSymptom   SymptomSource = "official"
	SuggestedSymptom  SymptomSource = "suggested"
	CustomizedSymptom SymptomSource = "customized"
)

const (
	SymptomCollection       = "symptom"
	SymptomReportCollection = "symptomReport"
)

type Symptom struct {
	ID     string        `json:"id" bson:"_id"`
	Name   string        `json:"name" bson:"name"`
	Desc   string        `json:"-" bson:"desc"`
	Source SymptomSource `json:"-" bson:"source"`
}

var (
	OfficialSymptoms = map[string]bool{
		"cough":            true,
		"breath":           true,
		"fever":            true,
		"chills":           true,
		"muscle_pain":      true,
		"throat":           true,
		"loss_taste_smell": true,
	}
)

// The system defined symptoms. The list will be inserted into database by migration function
var COVID19Symptoms = []Symptom{
	{ID: "cough", Name: "Cough", Source: OfficialSymptom},
	{ID: "breath", Name: "Shortness of breath or difficulty breathing", Source: OfficialSymptom},
	{ID: "fever", Name: "Fever", Source: OfficialSymptom},
	{ID: "chills", Name: "Chills", Source: OfficialSymptom},
	{ID: "muscle_pain", Name: "Muscle pain", Source: OfficialSymptom},
	{ID: "throat", Name: "Sore throat", Source: OfficialSymptom},
	{ID: "loss_taste_smell", Name: "New loss of taste or smell", Source: OfficialSymptom},
}

var GeneralSymptoms = []Symptom{
	{ID: "suggestion_1", Name: "Abdominal bloating", Source: SuggestedSymptom},
	{ID: "suggestion_2", Name: "Abdominal pain (stomachache)", Source: SuggestedSymptom},
	{ID: "suggestion_3", Name: "Abnormal amount of body hair growth (hypertrichosis)", Source: SuggestedSymptom},
	{ID: "suggestion_4", Name: "Abnormal skin tingling, prickling, chilling, burning, or numbness (paresthesia)", Source: SuggestedSymptom},
	{ID: "suggestion_5", Name: "Abnormal sweating (hyperhidrosis)", Source: SuggestedSymptom},
	{ID: "suggestion_6", Name: "Abnormal sweating (perspiration)", Source: SuggestedSymptom},
	{ID: "suggestion_7", Name: "Abnormal vaginal bleeding", Source: SuggestedSymptom},
	{ID: "suggestion_8", Name: "Abnormal walking (ataxia)", Source: SuggestedSymptom},
	{ID: "suggestion_9", Name: "Abnormally fast breathing (tachypnea)", Source: SuggestedSymptom},
	{ID: "suggestion_10", Name: "Abnormally slow breathing (bradypnea)", Source: SuggestedSymptom},
	{ID: "suggestion_11", Name: "Absence of menstrual period (amenorrhea)", Source: SuggestedSymptom},
	{ID: "suggestion_12", Name: "Anxiety", Source: SuggestedSymptom},
	{ID: "suggestion_13", Name: "Apathy", Source: SuggestedSymptom},
	{ID: "suggestion_14", Name: "Back pain (backache)", Source: SuggestedSymptom},
	{ID: "suggestion_15", Name: "Bad breath (halitosis)", Source: SuggestedSymptom},
	{ID: "suggestion_16", Name: "Blister", Source: SuggestedSymptom},
	{ID: "suggestion_17", Name: "Blood in semen (hematospermia)", Source: SuggestedSymptom},
	{ID: "suggestion_18", Name: "Blood in stool", Source: SuggestedSymptom},
	{ID: "suggestion_19", Name: "Blood in urine (hematuria)", Source: SuggestedSymptom},
	{ID: "suggestion_20", Name: "Blurred vision", Source: SuggestedSymptom},
	{ID: "suggestion_21", Name: "Bruising (contusions, hematoma)", Source: SuggestedSymptom},
	{ID: "suggestion_22", Name: "Burping (belching)", Source: SuggestedSymptom},
	{ID: "suggestion_23", Name: "Chills", Source: SuggestedSymptom},
	{ID: "suggestion_24", Name: "Chronic pain", Source: SuggestedSymptom},
	{ID: "suggestion_25", Name: "Confusion", Source: SuggestedSymptom},
	{ID: "suggestion_26", Name: "Convulsions", Source: SuggestedSymptom},
	{ID: "suggestion_27", Name: "Coughed up mucous (sputum)", Source: SuggestedSymptom},
	{ID: "suggestion_28", Name: "Coughing up blood (hemoptysis)", Source: SuggestedSymptom},
	{ID: "suggestion_29", Name: "Decreased appetite (anorexia)", Source: SuggestedSymptom},
	{ID: "suggestion_30", Name: "Deformity", Source: SuggestedSymptom},
	{ID: "suggestion_31", Name: "Delusions", Source: SuggestedSymptom},
	{ID: "suggestion_32", Name: "Depression", Source: SuggestedSymptom},
	{ID: "suggestion_33", Name: "Diarrhea", Source: SuggestedSymptom},
	{ID: "suggestion_34", Name: "Difficult or painful urination (dysuria)", Source: SuggestedSymptom},
	{ID: "suggestion_35", Name: "Difficulty swallowing (dysphagia)", Source: SuggestedSymptom},
	{ID: "suggestion_36", Name: "Dilated pupils", Source: SuggestedSymptom},
	{ID: "suggestion_37", Name: "Discharge of mucous or pus", Source: SuggestedSymptom},
	{ID: "suggestion_38", Name: "Dizziness (vertigo)", Source: SuggestedSymptom},
	{ID: "suggestion_39", Name: "Double vision", Source: SuggestedSymptom},
	{ID: "suggestion_40", Name: "Dry mouth (xerostomia)", Source: SuggestedSymptom},
	{ID: "suggestion_41", Name: "Ear pain (Earache, otalgia)", Source: SuggestedSymptom},
	{ID: "suggestion_42", Name: "Erectile dysfunction (impotence)", Source: SuggestedSymptom},
	{ID: "suggestion_43", Name: "Erratic heartbeat (heart palpitations)", Source: SuggestedSymptom},
	{ID: "suggestion_44", Name: "Excessive body hair (hirsutism)", Source: SuggestedSymptom},
	{ID: "suggestion_45", Name: "Excessive urination (polyuria)", Source: SuggestedSymptom},
	{ID: "suggestion_46", Name: "Eyelid spasms", Source: SuggestedSymptom},
	{ID: "suggestion_47", Name: "Fainting (loss of consciousness, syncope)", Source: SuggestedSymptom},
	{ID: "suggestion_48", Name: "Feeling of incomplete defecation (rectal tenesmus)", Source: SuggestedSymptom},
	{ID: "suggestion_49", Name: "Fingernail / toenail infection or deformity", Source: SuggestedSymptom},
	{ID: "suggestion_50", Name: "Flatulence (gas)", Source: SuggestedSymptom},
	{ID: "suggestion_51", Name: "Frequent urination", Source: SuggestedSymptom},
	{ID: "suggestion_52", Name: "Gastrointestinal bleeding (hemorrhaging)", Source: SuggestedSymptom},
	{ID: "suggestion_53", Name: "Hair loss (alopecia)", Source: SuggestedSymptom},
	{ID: "suggestion_54", Name: "Hallucination", Source: SuggestedSymptom},
	{ID: "suggestion_55", Name: "Headache", Source: SuggestedSymptom},
	{ID: "suggestion_56", Name: "Hearing loss", Source: SuggestedSymptom},
	{ID: "suggestion_57", Name: "Heart arrhythmia (cardiac arrhythmia)", Source: SuggestedSymptom},
	{ID: "suggestion_58", Name: "Heartburn (pyrosis)", Source: SuggestedSymptom},
	{ID: "suggestion_59", Name: "High resting heart rate (tachycardia)", Source: SuggestedSymptom},
	{ID: "suggestion_60", Name: "Inability or weakness moving one side of the body", Source: SuggestedSymptom},
	{ID: "suggestion_61", Name: "Inability to uninate normally (urinary retention)", Source: SuggestedSymptom},
	{ID: "suggestion_62", Name: "Indigestion (dyspepsia)", Source: SuggestedSymptom},
	{ID: "suggestion_63", Name: "Infertility", Source: SuggestedSymptom},
	{ID: "suggestion_64", Name: "Involuntary body movements", Source: SuggestedSymptom},
	{ID: "suggestion_65", Name: "Involuntary eye movements", Source: SuggestedSymptom},
	{ID: "suggestion_66", Name: "Involuntary urination (urinary incontinence)", Source: SuggestedSymptom},
	{ID: "suggestion_67", Name: "Itching", Source: SuggestedSymptom},
	{ID: "suggestion_68", Name: "Joint pain (arthralgia)", Source: SuggestedSymptom},
	{ID: "suggestion_69", Name: "Lightheadedness", Source: SuggestedSymptom},
	{ID: "suggestion_70", Name: "Loss of bowel control (fecal incontinence)", Source: SuggestedSymptom},
	{ID: "suggestion_71", Name: "Loss of normal bowel movements (constipation)", Source: SuggestedSymptom},
	{ID: "suggestion_72", Name: "Loss of normal speech (aphasia)", Source: SuggestedSymptom},
	{ID: "suggestion_73", Name: "Loss of sense of smell (anosmia)", Source: SuggestedSymptom},
	{ID: "suggestion_74", Name: "Loss of sense of taste (dysgeusia / parageusia)", Source: SuggestedSymptom},
	{ID: "suggestion_75", Name: "Loss of vision (blindness)", Source: SuggestedSymptom},
	{ID: "suggestion_76", Name: "Loss or writing ability (dysgraphia)", Source: SuggestedSymptom},
	{ID: "suggestion_77", Name: "Low body temperature (hypothermia)", Source: SuggestedSymptom},
	{ID: "suggestion_78", Name: "Low resting heart rate (bradycardia)", Source: SuggestedSymptom},
	{ID: "suggestion_79", Name: "Malaise", Source: SuggestedSymptom},
	{ID: "suggestion_80", Name: "Memory loss (amnesia)", Source: SuggestedSymptom},
	{ID: "suggestion_81", Name: "Muscle cramps", Source: SuggestedSymptom},
	{ID: "suggestion_82", Name: "Muscle loss (cachexia)", Source: SuggestedSymptom},
	{ID: "suggestion_83", Name: "Nausea", Source: SuggestedSymptom},
	{ID: "suggestion_84", Name: "Nervous tics", Source: SuggestedSymptom},
	{ID: "suggestion_85", Name: "Nosebleed (epistaxis)", Source: SuggestedSymptom},
	{ID: "suggestion_86", Name: "Oily or fatty feces (steatorrhea)", Source: SuggestedSymptom},
	{ID: "suggestion_87", Name: "Pain when swallowing (odynophagia)", Source: SuggestedSymptom},
	{ID: "suggestion_88", Name: "Painful intercourse", Source: SuggestedSymptom},
	{ID: "suggestion_89", Name: "Pelvic pain", Source: SuggestedSymptom},
	{ID: "suggestion_90", Name: "Rash", Source: SuggestedSymptom},
	{ID: "suggestion_91", Name: "Rectal or anal pain (proctologia fugax)", Source: SuggestedSymptom},
	{ID: "suggestion_92", Name: "Reduced ability to experience pleasure", Source: SuggestedSymptom},
	{ID: "suggestion_93", Name: "Reduced ability to open the jaws (lockjaw, trismus)", Source: SuggestedSymptom},
	{ID: "suggestion_94", Name: "Ringing or hissing in ears (tinnitus)", Source: SuggestedSymptom},
	{ID: "suggestion_95", Name: "Runny nose (rhinorrhea)", Source: SuggestedSymptom},
	{ID: "suggestion_96", Name: "Sciatica (pain going down the leg from lower back)", Source: SuggestedSymptom},
	{ID: "suggestion_97", Name: "Sharp chest pains while breathing (pleuritic chest pain)", Source: SuggestedSymptom},
	{ID: "suggestion_98", Name: "Shivering", Source: SuggestedSymptom},
	{ID: "suggestion_99", Name: "Skin pain (dermatome)", Source: SuggestedSymptom},
	{ID: "suggestion_100", Name: "Sleepiness (somnolence)", Source: SuggestedSymptom},
	{ID: "suggestion_101", Name: "Sleeplessness (insomnia)", Source: SuggestedSymptom},
	{ID: "suggestion_102", Name: "Sounds are too loud", Source: SuggestedSymptom},
	{ID: "suggestion_103", Name: "Stopping of breathing (apnea)", Source: SuggestedSymptom},
	{ID: "suggestion_104", Name: "Stopping of breathing while sleeping (sleep apnea)", Source: SuggestedSymptom},
	{ID: "suggestion_105", Name: "Swelling (edema)", Source: SuggestedSymptom},
	{ID: "suggestion_106", Name: "Swollen or painful lymph nodes (lymphadenopathy / adenopathy)", Source: SuggestedSymptom},
	{ID: "suggestion_107", Name: "Thirst", Source: SuggestedSymptom},
	{ID: "suggestion_108", Name: "Thoughts of killing yourself (suicidal ideation)", Source: SuggestedSymptom},
	{ID: "suggestion_109", Name: "Tooth pain (toothache)", Source: SuggestedSymptom},
	{ID: "suggestion_110", Name: "Tremors", Source: SuggestedSymptom},
	{ID: "suggestion_111", Name: "Urethral discharge", Source: SuggestedSymptom},
	{ID: "suggestion_112", Name: "Vaginal discharge", Source: SuggestedSymptom},
	{ID: "suggestion_113", Name: "Vomiting", Source: SuggestedSymptom},
	{ID: "suggestion_114", Name: "Vomiting blood (hematemesis)", Source: SuggestedSymptom},
	{ID: "suggestion_115", Name: "Weakness (muscle weakness)", Source: SuggestedSymptom},
	{ID: "suggestion_116", Name: "Weight gain", Source: SuggestedSymptom},
	{ID: "suggestion_117", Name: "Weight loss", Source: SuggestedSymptom},
	{ID: "suggestion_118", Name: "Wound (laceration)", Source: SuggestedSymptom},
	{ID: "suggestion_119", Name: "Yellowish or greenish skin (jaundice)", Source: SuggestedSymptom},
}

// SymptomReportData the struct to store symptom data and score
type SymptomReportData struct {
	ProfileID     string    `json:"profile_id" bson:"profile_id"`
	AccountNumber string    `json:"account_number" bson:"account_number"`
	Symptoms      []Symptom `json:"symptoms" bson:"symptoms"`
	Location      GeoJSON   `json:"location" bson:"location"`
	Timestamp     int64     `json:"ts" bson:"ts"`
}

type SymptomDistribution map[string]int

func (s *SymptomReportData) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Symptoms  []Symptom `json:"symptoms"`
		Location  Location  `json:"location"`
		Timestamp int64     `json:"timestamp"`
	}{
		Symptoms:  s.Symptoms,
		Location:  Location{Longitude: s.Location.Coordinates[0], Latitude: s.Location.Coordinates[1]},
		Timestamp: s.Timestamp,
	})
}

// SplitSymptoms separates official and non-official symptoms
func SplitSymptoms(symptoms []Symptom) ([]Symptom, []Symptom) {
	official := make([]Symptom, 0)
	nonOfficial := make([]Symptom, 0)
	for _, s := range symptoms {
		if OfficialSymptoms[s.ID] {
			official = append(official, s)
		} else {
			nonOfficial = append(nonOfficial, s)
		}
	}

	return official, nonOfficial
}
