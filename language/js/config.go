package js

import (
	"flag"

	"github.com/bazelbuild/bazel-gazelle/config"
	gzflag "github.com/bazelbuild/bazel-gazelle/flag"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

// jsConfig contains configuration values related to JS rules.
type jsConfig struct {
	// prefix is a prefix of an import path, used to generate importpath
	// attributes. Set with # gazelle:js_prefix.
	prefix string

	// prefixRel is the package name of the directory where the prefix was set
	// ("" for the root directory).
	prefixRel string

	// prefixSet indicates whether the prefix was set explicitly. It is an error
	// to infer an importpath for a rule without setting the prefix.
	prefixSet bool
}

func newJsConfig() *jsConfig {
	gc := &jsConfig{}
	return gc
}

func getJsConfig(c *config.Config) *jsConfig {
	return c.Exts[jsName].(*jsConfig)
}

func (gc *jsConfig) clone() *jsConfig {
	gcCopy := *gc
	return &gcCopy
}

func (_ *jsLang) KnownDirectives() []string {
	return []string{
		"js_prefix",
	}
}

func (_ *jsLang) RegisterFlags(fs *flag.FlagSet, cmd string, c *config.Config) {
	gc := newJsConfig()
	switch cmd {
	case "fix", "update":
		fs.Var(
			&gzflag.ExplicitFlag{Value: &gc.prefix, IsSet: &gc.prefixSet},
			"js_prefix",
			"prefix of import paths in the current workspace")

	}
	c.Exts[jsName] = gc
}

func (_ *jsLang) CheckFlags(fs *flag.FlagSet, c *config.Config) error {
	return nil
}

func (_ *jsLang) Configure(c *config.Config, rel string, f *rule.File) {
	var gc *jsConfig
	if raw, ok := c.Exts[jsName]; !ok {
		gc = newJsConfig()
	} else {
		gc = raw.(*jsConfig).clone()
	}
	c.Exts[jsName] = gc

	if f != nil {
		setPrefix := func(prefix string) {
			gc.prefix = prefix
			gc.prefixSet = true
			gc.prefixRel = rel
		}
		for _, d := range f.Directives {
			switch d.Key {
			case "js_prefix":
				setPrefix(d.Value)
			}
		}
		if !gc.prefixSet {
			for _, r := range f.Rules {
				switch r.Kind() {
				case "gazelle":
					if prefix := r.AttrString("js_prefix"); prefix != "" {
						setPrefix(prefix)
					}
				}
			}
		}
	}
}
