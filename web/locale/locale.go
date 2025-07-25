package locale

import (
	"embed"
	"io/fs"
	"strings"

	"x-ui/logger"

	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/text/language"
)

var (
	i18nBundle   *i18n.Bundle
	LocalizerWeb *i18n.Localizer
	LocalizerBot *i18n.Localizer
)

type I18nType string

const (
	Bot I18nType = "bot"
	Web I18nType = "web"
)



func InitLocalizer(i18nFS embed.FS) error {
	// set default bundle to english
	i18nBundle = i18n.NewBundle(language.MustParse("en-US"))
	i18nBundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	// parse files
	if err := parseTranslationFiles(i18nFS, i18nBundle); err != nil {
		return err
	}

	// Initialize localizers for web and bot
	LocalizerWeb = i18n.NewLocalizer(i18nBundle, "en-US")
	LocalizerBot = i18n.NewLocalizer(i18nBundle, "en-US")

	return nil
}

func createTemplateData(params []string, seperator ...string) map[string]any {
	var sep string = "=="
	if len(seperator) > 0 {
		sep = seperator[0]
	}

	templateData := make(map[string]any)
	for _, param := range params {
		parts := strings.SplitN(param, sep, 2)
		templateData[parts[0]] = parts[1]
	}

	return templateData
}

func I18n(i18nType I18nType, key string, params ...string) string {
	var localizer *i18n.Localizer

	switch i18nType {
	case "bot":
		localizer = LocalizerBot
	case "web":
		localizer = LocalizerWeb
	default:
		logger.Errorf("Invalid type for I18n: %s", i18nType)
		return ""
	}

	templateData := createTemplateData(params)

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: templateData,
	})
	if err != nil {
		logger.Errorf("Failed to localize message: %v", err)
		return ""
	}

	return msg
}



func LocalizerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Always use English
		LocalizerWeb = i18n.NewLocalizer(i18nBundle, "en-US")

		c.Set("localizer", LocalizerWeb)
		c.Set("I18n", I18n)
		c.Next()
	}
}

func parseTranslationFiles(i18nFS embed.FS, i18nBundle *i18n.Bundle) error {
	err := fs.WalkDir(i18nFS, "translation",
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			data, err := i18nFS.ReadFile(path)
			if err != nil {
				return err
			}

			_, err = i18nBundle.ParseMessageFileBytes(data, path)
			return err
		})
	if err != nil {
		return err
	}

	return nil
}
