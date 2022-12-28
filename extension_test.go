package gazelle_ext

import (
	"reflect"
	"testing"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
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
		got := lang.GenerateRules(language.GenerateArgs{RegularFiles: []string{"test.java"}})

		want := language.GenerateResult{}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %#v, want %+v", got, want)
		}

	})

	t.Run("resolve", func(t *testing.T) {
		ret := lang.GenerateRules(language.GenerateArgs{RegularFiles: []string{"test.java"}})

		for i, r := range ret.Gen {
			lang.Resolve(config.New().Clone(), &resolve.RuleIndex{}, &repo.RemoteCache{}, r, ret.Imports[i], label.New("", "", r.Name()))

		}
	})

}
