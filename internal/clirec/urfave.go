// This file adapts urfave/cli v3.

package clirec

import (
	"context"

	cli "github.com/urfave/cli/v3"
)

// WrapUrfave returns the flags wrapped so that every command-line occurrence is
// added to rec. Values that reach a flag through its Sources (env, file) are
// not recorded: urfave applies them on the inner flag, not through the wrapper.
func WrapUrfave(rec *Recorder, flags []cli.Flag) []cli.Flag {
	out := make([]cli.Flag, len(flags))
	for i, f := range flags {
		out[i] = &flag{Flag: f, rec: rec}
	}
	return out
}

// flag embeds the inner flag, so every method of the cli.Flag interface is
// promoted. urfave also type-asserts optional interfaces; embedding an
// interface hides them, so each one is forwarded below.
type flag struct {
	cli.Flag
	rec *Recorder
}

// Set is the one override: urfave calls it once per command-line occurrence,
// in command-line order, with the name as typed.
func (f *flag) Set(name, value string) error {
	if err := f.Flag.Set(name, value); err != nil {
		return err
	}
	f.rec.Add(Occurrence{Flag: f.Flag.Names()[0], Typed: name, Value: value})
	return nil
}

// Forwarded optional interfaces. Without IsBoolFlag a bool flag would take the
// next argument as its value.

func (f *flag) IsBoolFlag() bool {
	b, ok := f.Flag.(interface{ IsBoolFlag() bool })
	return ok && b.IsBoolFlag()
}

func (f *flag) IsLocal() bool {
	l, ok := f.Flag.(cli.LocalFlag)
	return ok && l.IsLocal()
}

func (f *flag) IsRequired() bool {
	r, ok := f.Flag.(cli.RequiredFlag)
	return ok && r.IsRequired()
}

func (f *flag) IsVisible() bool {
	v, ok := f.Flag.(cli.VisibleFlag)
	return !ok || v.IsVisible()
}

func (f *flag) GetCategory() string {
	if c, ok := f.Flag.(cli.CategorizableFlag); ok {
		return c.GetCategory()
	}
	return ""
}

func (f *flag) SetCategory(s string) {
	if c, ok := f.Flag.(cli.CategorizableFlag); ok {
		c.SetCategory(s)
	}
}

func (f *flag) Count() int {
	if c, ok := f.Flag.(cli.Countable); ok {
		return c.Count()
	}
	return 0
}

func (f *flag) RunAction(ctx context.Context, cmd *cli.Command) error {
	if a, ok := f.Flag.(cli.ActionableFlag); ok {
		return a.RunAction(ctx, cmd)
	}
	return nil
}

func (f *flag) doc() cli.DocGenerationFlag {
	d, _ := f.Flag.(cli.DocGenerationFlag)
	return d
}

func (f *flag) TakesValue() bool       { return f.doc() != nil && f.doc().TakesValue() }
func (f *flag) GetUsage() string       { return f.doc().GetUsage() }
func (f *flag) GetValue() string       { return f.doc().GetValue() }
func (f *flag) GetDefaultText() string { return f.doc().GetDefaultText() }
func (f *flag) GetEnvVars() []string   { return f.doc().GetEnvVars() }
func (f *flag) IsDefaultVisible() bool { return f.doc().IsDefaultVisible() }
func (f *flag) TypeName() string       { return f.doc().TypeName() }

func (f *flag) IsMultiValueFlag() bool {
	m, ok := f.Flag.(cli.DocGenerationMultiValueFlag)
	return ok && m.IsMultiValueFlag()
}
