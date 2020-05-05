package utils

import (
	"path"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

var bundle *i18n.Bundle

func InitI18NBundle() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	bundle.MustLoadMessageFile(path.Join(viper.GetString("i18n.dir"), "en.yaml"))
	bundle.MustLoadMessageFile(path.Join(viper.GetString("i18n.dir"), "zh_tw.yaml"))
}

func NewLocalizer(lang string) *i18n.Localizer {
	return i18n.NewLocalizer(bundle, lang)
}
