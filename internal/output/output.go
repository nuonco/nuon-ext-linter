package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/nuonco/nuon-ext-linter/internal/rule"
)

func PrintText(w io.Writer, findings []rule.Finding) {
	if len(findings) == 0 {
		fmt.Fprintln(w, "No issues found.")
		return
	}

	for _, f := range findings {
		severity := strings.ToUpper(f.Severity.String())
		if f.File != "" {
			fmt.Fprintf(w, "  %s  %s  %s\n", severity, f.File, f.Message)
		} else {
			fmt.Fprintf(w, "  %s  %s\n", severity, f.Message)
		}
	}

	errors, warnings, infos := countBySeverity(findings)
	fmt.Fprintf(w, "\n%d error(s), %d warning(s), %d info(s)\n", errors, warnings, infos)
}

func PrintJSON(w io.Writer, findings []rule.Finding) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(map[string]any{
		"findings": findings,
	})
}

func countBySeverity(findings []rule.Finding) (errors, warnings, infos int) {
	for _, f := range findings {
		switch f.Severity {
		case rule.SeverityError:
			errors++
		case rule.SeverityWarning:
			warnings++
		case rule.SeverityInfo:
			infos++
		}
	}
	return
}
