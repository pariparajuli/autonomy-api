package background

import (
	"fmt"
	"strings"

	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/bitmark-inc/autonomy-api/utils"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// CommaSeparatedSymptoms will return a string of symptoms separate by commas
func CommaSeparatedSymptoms(lang string, sourceSymptoms []schema.Symptom) string {
	loc := utils.NewLocalizer(lang)

	symptomsNames := make([]string, 0)

	for _, s := range sourceSymptoms {
		if name, err := loc.Localize(&i18n.LocalizeConfig{
			MessageID: fmt.Sprintf("symptoms.%s.name", s.ID),
		}); err == nil {
			symptomsNames = append(symptomsNames, name)
		} else {
			symptomsNames = append(symptomsNames, s.Name)
		}
	}

	return strings.Join(symptomsNames, ", ")
}
