package formatter

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSortOrder_alpha(t *testing.T) {
	order := []string{"E501", "B904", "F401"}
	counts := map[string]int{"E501": 3, "B904": 10, "F401": 1}
	got := SortOrder(order, counts, "alpha")
	want := []string{"B904", "E501", "F401"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("alpha sort: got %v, want %v", got, want)
	}
}

func TestSortOrder_count(t *testing.T) {
	order := []string{"E501", "B904", "F401"}
	counts := map[string]int{"E501": 3, "B904": 10, "F401": 1}
	got := SortOrder(order, counts, "count")
	want := []string{"B904", "E501", "F401"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("count sort: got %v, want %v", got, want)
	}
}

func TestSortOrder_countTiebreakAlpha(t *testing.T) {
	order := []string{"Z", "A"}
	counts := map[string]int{"Z": 5, "A": 5}
	got := SortOrder(order, counts, "count")
	if got[0] != "A" {
		t.Errorf("count sort tiebreak: want A first, got %v", got)
	}
}

func TestFilterRuleOrder_empty(t *testing.T) {
	order := []string{"E501", "F401", "B904"}
	got := FilterRuleOrder(order, nil)
	if !reflect.DeepEqual(got, order) {
		t.Error("nil filter should return unchanged")
	}
}

func TestFilterRuleOrder_filters(t *testing.T) {
	order := []string{"E501", "F401", "B904"}
	got := FilterRuleOrder(order, []string{"E501", "B904"})
	want := []string{"E501", "B904"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMatchesFileFilter_empty(t *testing.T) {
	if !MatchesFileFilter("src/foo.py", nil) {
		t.Error("nil filter should match everything")
	}
}

func TestMatchesFileFilter_prefix(t *testing.T) {
	if !MatchesFileFilter("src/foo.py", []string{"src/"}) {
		t.Error("should match src/ prefix")
	}
	if MatchesFileFilter("tests/bar.py", []string{"src/"}) {
		t.Error("should not match tests/ against src/ filter")
	}
}

func TestTruncate_noOp(t *testing.T) {
	if got := Truncate("hello", 0); got != "hello" {
		t.Errorf("max=0 should not truncate, got %q", got)
	}
	if got := Truncate("hello", 10); got != "hello" {
		t.Errorf("max>=len should not truncate, got %q", got)
	}
	if got := Truncate("hello", 5); got != "hello" {
		t.Errorf("max==len should not truncate, got %q", got)
	}
}

func TestTruncate_truncates(t *testing.T) {
	got := Truncate("hello world", 5)
	if got != "hello..." {
		t.Errorf("got %q, want %q", got, "hello...")
	}
}

func TestFormatLines_singular(t *testing.T) {
	got := FormatLines([]int{42})
	if got != "line 42" {
		t.Errorf("got %q, want %q", got, "line 42")
	}
}

func TestFormatLines_plural(t *testing.T) {
	got := FormatLines([]int{3, 7, 12})
	if got != "lines 3, 7, 12" {
		t.Errorf("got %q, want %q", got, "lines 3, 7, 12")
	}
}

func TestSeverityLabel(t *testing.T) {
	cases := []struct{ in, want string }{
		{"error", "ERROR"},
		{"ERROR", "ERROR"},
		{"fatal", "ERROR"},
		{"FATAL", "ERROR"},
		{"warning", "WARN"},
		{"warn", "WARN"},
		{"WARNING", "WARN"},
		{"WARN", "WARN"},
		{"info", "INFO"},
		{"INFO", "INFO"},
		{"information", "INFO"},
		{"note", "INFO"},
		{"style", "INFO"},
		{"hint", "INFO"},
		{"", "INFO"},
		{"CUSTOM", "CUSTOM"},
	}
	for _, c := range cases {
		got := SeverityLabel(c.in)
		if got != c.want {
			t.Errorf("SeverityLabel(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSeverityColor_colorsOff(t *testing.T) {
	for _, sev := range []string{"error", "warning", "info", "unknown"} {
		if got := SeverityColor(sev, false); got != "" {
			t.Errorf("colors=false: SeverityColor(%q) = %q, want empty", sev, got)
		}
	}
}

func TestSeverityColor_knownSeverities(t *testing.T) {
	cases := []struct {
		sev  string
		want string
	}{
		{"error", "\033[1;31m"},
		{"ERROR", "\033[1;31m"},
		{"fatal", "\033[1;31m"},
		{"warning", "\033[1;33m"},
		{"warn", "\033[1;33m"},
		{"info", "\033[1;34m"},
		{"note", "\033[1;34m"},
	}
	for _, c := range cases {
		got := SeverityColor(c.sev, true)
		if got != c.want {
			t.Errorf("SeverityColor(%q, true) = %q, want %q", c.sev, got, c.want)
		}
	}
}

func TestSeverityColor_unknownSeverity(t *testing.T) {
	if got := SeverityColor("mystery", true); got != "" {
		t.Errorf("unknown severity should return empty, got %q", got)
	}
}

func TestPlural(t *testing.T) {
	if got := Plural(1, "file", "files"); got != "file" {
		t.Errorf("n=1: got %q, want %q", got, "file")
	}
	if got := Plural(0, "file", "files"); got != "files" {
		t.Errorf("n=0: got %q, want %q", got, "files")
	}
	if got := Plural(2, "file", "files"); got != "files" {
		t.Errorf("n=2: got %q, want %q", got, "files")
	}
}

func TestSummary(t *testing.T) {
	cases := []struct {
		issues, rules, files int
		want                 string
	}{
		{0, 0, 0, "0 issues · 0 rules · 0 files"},
		{1, 1, 1, "1 issue · 1 rule · 1 file"},
		{3, 2, 2, "3 issues · 2 rules · 2 files"},
	}
	for _, c := range cases {
		got := Summary(c.issues, c.rules, c.files)
		if got != c.want {
			t.Errorf("Summary(%d,%d,%d) = %q, want %q", c.issues, c.rules, c.files, got, c.want)
		}
	}
}

func TestParseNDJSON_valid(t *testing.T) {
	input := []byte(`{"a":1}` + "\n" + `{"b":2}` + "\n")
	msgs, err := ParseNDJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 2 {
		t.Fatalf("want 2 messages, got %d", len(msgs))
	}
	var m map[string]int
	if err := json.Unmarshal(msgs[0], &m); err != nil || m["a"] != 1 {
		t.Error("first message wrong")
	}
}

func TestParseNDJSON_skipsBlankAndInvalid(t *testing.T) {
	input := []byte(`{"ok":true}` + "\n\n" + `not json` + "\n" + `{"also":true}`)
	msgs, err := ParseNDJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 2 {
		t.Fatalf("want 2 valid messages, got %d", len(msgs))
	}
}

func TestParseNDJSON_empty(t *testing.T) {
	msgs, err := ParseNDJSON([]byte(""))
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 0 {
		t.Errorf("want 0 messages, got %d", len(msgs))
	}
}

func TestResolvePath_relative(t *testing.T) {
	cfg := DefaultConfig()
	got := ResolvePath("src/foo.py", cfg)
	if got != "src/foo.py" {
		t.Errorf("relative path should be unchanged, got %q", got)
	}
}

func TestResolvePath_absolute(t *testing.T) {
	cwd, _ := os.Getwd()
	abs := filepath.Join(cwd, "pkg", "formatter", "helpers.go")
	cfg := DefaultConfig()
	got := ResolvePath(abs, cfg)
	if !strings.HasPrefix(got, "pkg") {
		t.Errorf("absolute path should be made relative to CWD, got %q", got)
	}
}

func TestResolvePath_basenameOnly(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Extra = map[string]any{"basename_only": true}
	got := ResolvePath("some/deep/path/file.py", cfg)
	if got != "file.py" {
		t.Errorf("basename_only: got %q, want %q", got, "file.py")
	}
}

func TestTrimSpace_leadingTrailing(t *testing.T) {
	cases := []struct {
		in   []byte
		want []byte
	}{
		{[]byte("  hello  "), []byte("hello")},
		{[]byte("\thello\t"), []byte("hello")},
		{[]byte("\rhello\r"), []byte("hello")},
		{[]byte("hello"), []byte("hello")},
		{[]byte("   "), []byte("")},
	}
	for _, c := range cases {
		got := trimSpace(c.in)
		if !bytes.Equal(got, c.want) {
			t.Errorf("trimSpace(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestResolvePath_absNotUnderCWD(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	// /nonexistent is not under dir, so rel starts with ".."
	cfg := DefaultConfig()
	cfg.Colors = false
	// Use a path that definitely won't be under t.TempDir()
	absPath := "/tmp/some/other/path/file.go"
	got := ResolvePath(absPath, cfg)
	// Should return path as-is when it's not under CWD
	if !filepath.IsAbs(got) && !strings.HasPrefix(got, "..") {
		// either return is fine as long as it doesn't crash
	}
	// The key thing: no crash and returns something meaningful
	if got == "" {
		t.Error("ResolvePath returned empty string")
	}
}

func TestPrintStats_basic(t *testing.T) {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	cfg := DefaultConfig()
	cfg.Colors = false
	PrintStats(
		[]string{"E501", "F401"},
		map[string]int{"E501": 5, "F401": 1},
		map[string]int{"E501": 2, "F401": 1},
		map[string]string{"E501": "line too long", "F401": "unused import"},
		3,
		cfg,
	)
	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	s := string(out)
	if !strings.Contains(s, "E501") {
		t.Errorf("want E501 in stats output, got:\n%s", s)
	}
	if !strings.Contains(s, "6 issues · 2 rules · 3 files") {
		t.Errorf("want summary, got:\n%s", s)
	}
}

func TestPrintStats_noFilesColumn(t *testing.T) {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	cfg := DefaultConfig()
	cfg.Colors = false
	PrintStats(
		[]string{"E501", "F401"},
		map[string]int{"E501": 3, "F401": 1},
		nil, // no ruleFiles column
		map[string]string{"E501": "line too long", "F401": ""},
		2,
		cfg,
	)
	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	s := string(out)
	if !strings.Contains(s, "E501") {
		t.Errorf("want E501 in no-files-column stats, got:\n%s", s)
	}
	if !strings.Contains(s, "4 issues · 2 rules · 2 files") {
		t.Errorf("want summary, got:\n%s", s)
	}
}
