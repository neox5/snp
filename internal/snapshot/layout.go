package snapshot

func (s *Snapshot) buildLayout() {
	var layout []Content
	needsDivider := false

	section := func(contents ...Content) {
		if needsDivider {
			layout = append(layout, &divider{})
		}
		layout = append(layout, contents...)
		needsDivider = true
	}

	if !s.Config.NoSummary {
		section(s.Summary)
	}

	if !s.Config.NoIndex {
		section(&header{Text: "File Index"}, &index{Entries: s.Entries})
	}

	if s.GitData != nil && !s.Config.NoGitLog {
		section(&header{Text: "Git Log (git adog)"}, s.GitData)
	}

	if !s.Config.NoContent {
		for _, e := range s.Entries {
			if e.IsDir {
				continue
			}
			section(&header{Text: e.RelPath}, e)
		}
	}

	// ### pass 2 — line calculation
	currentLine := 1
	for _, c := range layout {
		if e, ok := c.(*Entry); ok {
			e.StartLine = currentLine
			e.EndLine = currentLine + e.LineCount() - 1
		}
		currentLine += c.LineCount()
	}

	s.Summary.TotalLines = currentLine - 1
	s.Layout = layout
}
