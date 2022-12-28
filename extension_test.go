package gazelle_ext

import (
	"testing"

	"github.com/bazelbuild/bazel-gazelle/language"
)

func TestExtension(t *testing.T) {
	lang := NewLanguage()

	t.Run("language name", func(t *testing.T) {
		got := lang.Name()
		want := "java"
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})

	t.Run("parse", func(t *testing.T) {
		lang.GenerateRules(language.GenerateArgs{})
	})

}
