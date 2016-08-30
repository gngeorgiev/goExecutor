package languages

import (
	"fmt"

	"github.com/gngeorgiev/goExecutor/languages/baseLanguage"
	"github.com/gngeorgiev/goExecutor/languages/javascript"
	"github.com/go-errors/errors"
)

const (
	LanguageNameJavascript = "js"
)

func GetLanguage(language string) (baseLanguage.Language, error) {
	if language == LanguageNameJavascript {
		return javascript.JavascriptLanguage{}, nil
	}

	return nil, errors.New(fmt.Sprintf("Language not found %s", language))
}
