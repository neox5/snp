package clirec

// Occurrence is one appearance of a flag on the command line.
type Occurrence struct {
	// Flag is the canonical flag name ("output" even when "-o" was typed).
	Flag string
	// Typed is the name exactly as typed, without dashes.
	Typed string
	// Value is the raw text as given. A bare boolean flag carries "true".
	Value string
}

// Recorder collects occurrences in command-line order. The zero value is ready
// to use.
type Recorder struct {
	list []Occurrence
}

// Add appends an occurrence.
func (r *Recorder) Add(o Occurrence) {
	r.list = append(r.list, o)
}

// Occurrences returns a copy of the recorded occurrences in order.
func (r *Recorder) Occurrences() []Occurrence {
	return append([]Occurrence(nil), r.list...)
}
