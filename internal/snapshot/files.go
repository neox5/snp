package snapshot

import (
	"fmt"
	"io"
	"os"
)

func appendFile(w io.Writer, fullPath, relPath string) error {
	if _, err := fmt.Fprintf(w, "# %s\n", relPath); err != nil {
		return err
	}

	f, err := os.Open(fullPath)
	if err != nil {
		return fmt.Errorf("cannot open %q: %w", fullPath, err)
	}
	defer f.Close()

	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("cannot copy %q: %w", fullPath, err)
	}

	if _, err := io.WriteString(w, "\n\n"); err != nil {
		return err
	}

	return nil
}
