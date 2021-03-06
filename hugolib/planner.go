package hugolib

import (
	"fmt"
	"io"
)

func (s *Site) ShowPlan(out io.Writer) (err error) {
	if s.Source == nil || len(s.Source.Files()) <= 0 {
		fmt.Fprintf(out, "No source files provided.\n")
	}

	for _, p := range s.Pages {
		fmt.Fprintf(out, "%s", p.FileName)
		if p.IsRenderable() {
			fmt.Fprintf(out, " (renderer: markdown)")
		} else {
			fmt.Fprintf(out, " (renderer: n/a)")
		}
		if s.Tmpl != nil {
			fmt.Fprintf(out, " (layout: %s, exists: %t)", p.Layout(), s.Tmpl.Lookup(p.Layout()) != nil)
		}
		fmt.Fprintf(out, "\n")
		fmt.Fprintf(out, " canonical => ")
		if s.Target == nil {
			fmt.Fprintf(out, "%s\n\n", "!no target specified!")
			continue
		}

		trns, err := s.Target.Translate(p.TargetPath())
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s\n", trns)

		if s.Alias == nil {
			continue
		}

		for _, alias := range p.Aliases {
			aliasTrans, err := s.Alias.Translate(alias)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, " %s => %s\n", alias, aliasTrans)
		}
		fmt.Fprintln(out)
	}
	return
}
