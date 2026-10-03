package config

import (
	"strings"
	"time"
)

// Origins maps a setting's JSON key to the sources that contributed to it.
// Scalars name the last source that stated them; lists name every source that
// added entries, joined by "+".
type Origins map[string]string

// Resolved is the result of loading all layers.
type Resolved struct {
	// Config is the effective configuration for this run.
	Config *Config
	// Origins records where each effective setting came from.
	Origins Origins
	// Persist is the file and CLI layers folded without defaults; it is what
	// --save-config writes.
	Persist Layer
}

// fold resolves layers in order, later layers winning. Scalars take the last
// stated value. MatcherFlags append in order. Pick and force lists are unions.
func fold(sources []Source) (Layer, Origins) {
	var out Layer
	o := Origins{}

	for _, s := range sources {
		l := s.Layer

		o.setInt(&out.Depth, l.Depth, "depth", s.Name)
		o.setString(&out.OutputPath, l.OutputPath, "output_path", s.Name)
		o.setBool(&out.NoSummary, l.NoSummary, "no_summary", s.Name)
		o.setBool(&out.NoIndex, l.NoIndex, "no_index", s.Name)
		o.setBool(&out.NoGitLog, l.NoGitLog, "no_git_log", s.Name)
		o.setBool(&out.NoContent, l.NoContent, "no_content", s.Name)
		o.setBool(&out.Stdout, l.Stdout, "stdout", s.Name)
		o.setBool(&out.Silent, l.Silent, "silent", s.Name)

		if len(l.MatcherFlags) > 0 {
			for _, f := range l.MatcherFlags {
				f.Source = s.Name
				out.MatcherFlags = append(out.MatcherFlags, f)
			}
			o.addList("filter_flags", s.Name)
		}
		if len(l.PickPaths) > 0 {
			out.PickPaths = mergeUnique(out.PickPaths, l.PickPaths)
			o.addList("pick_paths", s.Name)
		}
		if len(l.ForceTextPatterns) > 0 {
			out.ForceTextPatterns = mergeUnique(out.ForceTextPatterns, l.ForceTextPatterns)
			o.addList("force_text_patterns", s.Name)
		}
		if len(l.ForceBinaryPatterns) > 0 {
			out.ForceBinaryPatterns = mergeUnique(out.ForceBinaryPatterns, l.ForceBinaryPatterns)
			o.addList("force_binary_patterns", s.Name)
		}
	}

	return out, o
}

func (o Origins) setInt(dst **int, src *int, key, name string) {
	if src == nil {
		return
	}
	v := *src
	*dst = &v
	o[key] = name
}

func (o Origins) setString(dst **string, src *string, key, name string) {
	if src == nil {
		return
	}
	v := *src
	*dst = &v
	o[key] = name
}

func (o Origins) setBool(dst **bool, src *bool, key, name string) {
	if src == nil {
		return
	}
	v := *src
	*dst = &v
	o[key] = name
}

func (o Origins) addList(key, name string) {
	prev := o[key]
	if prev == "" {
		o[key] = name
		return
	}
	for _, p := range strings.Split(prev, "+") {
		if p == name {
			return
		}
	}
	o[key] = prev + "+" + name
}

// toConfig turns a folded layer into a Config. Unstated settings take their
// zero value; DefaultLayer states Depth and OutputPath, so they are always set
// when defaults are part of the fold.
func toConfig(l Layer) Config {
	c := Config{
		Depth:               -1,
		MatcherFlags:        l.MatcherFlags,
		PickPaths:           l.PickPaths,
		ForceTextPatterns:   l.ForceTextPatterns,
		ForceBinaryPatterns: l.ForceBinaryPatterns,
	}
	if l.Depth != nil {
		c.Depth = *l.Depth
	}
	if l.OutputPath != nil {
		c.OutputPath = *l.OutputPath
	}
	c.NoSummary = boolOf(l.NoSummary)
	c.NoIndex = boolOf(l.NoIndex)
	c.NoGitLog = boolOf(l.NoGitLog)
	c.NoContent = boolOf(l.NoContent)
	c.Stdout = boolOf(l.Stdout)
	c.Silent = boolOf(l.Silent)
	return c
}

func boolOf(p *bool) bool {
	return p != nil && *p
}

// Load resolves defaults, the config file in p.SourceDir (unless p.NoConfig)
// and the CLI layer, in that order.
func Load(p Params, cli Layer) (*Resolved, error) {
	user := []Source{}

	if !p.NoConfig {
		file, found, err := loadLayer(p.SourceDir)
		if err != nil {
			return nil, err
		}
		if found {
			user = append(user, Source{Name: SourceFile, Layer: file})
		}
	}
	user = append(user, Source{Name: SourceCLI, Layer: cli})

	all := append([]Source{{Name: SourceDefault, Layer: DefaultLayer()}}, user...)
	eff, origins := fold(all)
	persist, _ := fold(user)

	cfg := toConfig(eff)
	cfg.Generated = time.Now()
	cfg.SourceDir = p.SourceDir
	cfg.DryRun = p.DryRun

	return &Resolved{Config: &cfg, Origins: origins, Persist: persist}, nil
}

func mergeUnique(base, extra []string) []string {
	seen := make(map[string]bool, len(base)+len(extra))
	var result []string
	for _, v := range append(append([]string(nil), base...), extra...) {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}
