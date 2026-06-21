package formatter

import (
	"fmt"
	"sort"
)

// Issue represents a single finding from any linter or tool.
type Issue struct {
	Rule     string // machine key: used for grouping, filtering, sorting
	Display  string // label shown in output; falls back to Rule if empty
	File     string // already resolved (caller runs ResolvePath before mapping)
	Line     int
	Message  string
	Severity string // "error"/"warning"/"info" variants -> SeverityColor/SeverityLabel
	Note     string // optional extra in rule header, e.g. "HIGH/MEDIUM" for bandit
}

// Render dispatches to RenderByRule or RenderByFile based on cfg.GroupBy.
func Render(issues []Issue, cfg Config) error {
	if cfg.GroupBy == "file" {
		return RenderByFile(issues, cfg)
	}
	return RenderByRule(issues, cfg)
}

// RenderByRule groups issues by rule and prints them.
func RenderByRule(issues []Issue, cfg Config) error {
	type ruleData struct {
		display  string
		message  string
		severity string
		note     string
		files    map[string]map[int]struct{}
		order    []string // insertion order of files
	}

	rules := map[string]*ruleData{}
	var ruleOrder []string
	distinctFiles := map[string]struct{}{}

	// 1. Build buckets
	for _, iss := range issues {
		rule := iss.Rule
		if rule == "" {
			rule = "UNKNOWN"
		}
		
		distinctFiles[iss.File] = struct{}{}

		if _, ok := rules[rule]; !ok {
			rules[rule] = &ruleData{
				display:  iss.Display,
				message:  iss.Message,
				severity: iss.Severity,
				note:     iss.Note,
				files:    make(map[string]map[int]struct{}),
			}
			if rules[rule].display == "" {
				rules[rule].display = rule
			}
			ruleOrder = append(ruleOrder, rule)
		}

		rd := rules[rule]
		if _, ok := rd.files[iss.File]; !ok {
			rd.files[iss.File] = make(map[int]struct{})
			rd.order = append(rd.order, iss.File)
		}
		rd.files[iss.File][iss.Line] = struct{}{}
	}

	// 2. Count + filter + sort
	ruleCounts := make(map[string]int, len(ruleOrder))
	for _, rule := range ruleOrder {
		count := 0
		for _, lines := range rules[rule].files {
			count += len(lines)
		}
		ruleCounts[rule] = count
	}

	ruleOrder = FilterRuleOrder(ruleOrder, cfg.OnlyRules)
	ruleOrder = SortOrder(ruleOrder, ruleCounts, cfg.Sort)

	// 3. Stats mode
	if cfg.Stats {
		ruleFiles := make(map[string]int, len(ruleOrder))
		ruleMessages := make(map[string]string, len(ruleOrder))
		for _, rule := range ruleOrder {
			ruleFiles[rule] = len(rules[rule].files)
			ruleMessages[rule] = rules[rule].message
		}
		PrintStats(ruleOrder, ruleCounts, ruleFiles, ruleMessages, len(distinctFiles), cfg)
		return nil
	}

	// 4. Print each rule
	displayedIssues := 0
	for _, rule := range ruleOrder {
		rd := rules[rule]

		hasMatchingFile := false
		for _, file := range rd.order {
			if MatchesFileFilter(file, cfg.OnlyFiles) {
				hasMatchingFile = true
				break
			}
		}
		if !hasMatchingFile {
			continue
		}

		count := ruleCounts[rule]
		displayedIssues += count

		colorPrefix := SeverityColor(rd.severity, cfg.Colors)
		colorSuffix := ""
		if colorPrefix != "" {
			colorSuffix = "\033[0m"
		}

		header := ""
		displayVal := rd.display
		if rd.severity != "" {
			header += fmt.Sprintf("[%s] ", SeverityLabel(rd.severity))
		}
		header += displayVal
		if rd.note != "" {
			header += fmt.Sprintf(" (%s)", rd.note)
		}
		header += fmt.Sprintf(" (%d)", count)
		if rd.message != "" {
			if cfg.Colors {
				header += fmt.Sprintf(" — \033[1m%s\033[0m", rd.message)
			} else {
				header += fmt.Sprintf(" — %s", rd.message)
			}
		}

		fmt.Printf("%s%s%s\n", colorPrefix, header, colorSuffix)
		fmt.Println("Affected files:")

		sortedFiles := make([]string, len(rd.order))
		copy(sortedFiles, rd.order)
		sort.Strings(sortedFiles)

		for _, file := range sortedFiles {
			if !MatchesFileFilter(file, cfg.OnlyFiles) {
				continue
			}
			linesMap := rd.files[file]
			var lines []int
			for line := range linesMap {
				if line > 0 {
					lines = append(lines, line)
				}
			}
			sort.Ints(lines)
			
			if len(lines) == 0 {
				fmt.Printf("  - %s\n", file)
			} else {
				fmt.Printf("  - %s — %s\n", file, FormatLines(lines))
			}
		}
		fmt.Println(Divider)
	}

	// 5. Summary
	fmt.Println(Summary(displayedIssues, len(ruleOrder), len(distinctFiles)))
	return nil
}

// RenderByFile groups issues by file and prints them.
func RenderByFile(issues []Issue, cfg Config) error {
	type lineEntry struct {
		rule    string
		display string
		line    int
		message string
	}

	files := map[string][]lineEntry{}
	var fileOrder []string
	
	// Track distinct rules for the summary
	distinctRules := map[string]struct{}{}

	// 1. Build buckets
	for _, iss := range issues {
		rule := iss.Rule
		if rule == "" {
			rule = "UNKNOWN"
		}
		distinctRules[rule] = struct{}{}
		
		display := iss.Display
		if display == "" {
			display = rule
		}

		if _, ok := files[iss.File]; !ok {
			files[iss.File] = []lineEntry{}
			fileOrder = append(fileOrder, iss.File)
		}

		// Deduplicate (rule, line) per file
		duplicate := false
		for _, existing := range files[iss.File] {
			if existing.rule == rule && existing.line == iss.Line {
				duplicate = true
				break
			}
		}
		
		if !duplicate {
			files[iss.File] = append(files[iss.File], lineEntry{
				rule:    rule,
				display: display,
				line:    iss.Line,
				message: iss.Message,
			})
		}
	}

	// 2. Filter + sort files
	filteredFiles := fileOrder[:0:0]
	for _, f := range fileOrder {
		if MatchesFileFilter(f, cfg.OnlyFiles) {
			filteredFiles = append(filteredFiles, f)
		}
	}
	fileOrder = filteredFiles
	sort.Strings(fileOrder)

	// 3. Print each file
	totalIssues := 0
	for _, file := range fileOrder {
		entries := files[file]
		
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].line != entries[j].line {
				return entries[i].line < entries[j].line
			}
			return entries[i].rule < entries[j].rule
		})

		var filteredEntries []lineEntry
		for _, e := range entries {
			if len(cfg.OnlyRules) > 0 {
				found := false
				for _, r := range cfg.OnlyRules {
					if e.rule == r {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			filteredEntries = append(filteredEntries, e)
		}
		
		if len(filteredEntries) == 0 {
			continue
		}
		
		totalIssues += len(filteredEntries)

		fmt.Printf("%s — %d %s\n", file, len(filteredEntries), Plural(len(filteredEntries), "issue", "issues"))

		lastRule := ""
		for _, e := range filteredEntries {
			msg := ""
			if e.rule != lastRule {
				if e.message != "" {
					msg = "  — " + e.message
				}
				lastRule = e.rule
			}
			
			if e.line > 0 {
				if cfg.Colors {
					fmt.Printf("  \033[1;33m%s\033[0m  line %d%s\n", e.display, e.line, msg)
				} else {
					fmt.Printf("  %s  line %d%s\n", e.display, e.line, msg)
				}
			} else {
				if cfg.Colors {
					fmt.Printf("  \033[1;33m%s\033[0m%s\n", e.display, msg)
				} else {
					fmt.Printf("  %s%s\n", e.display, msg)
				}
			}
		}
		fmt.Println(Divider)
	}

	// 4. Summary
	fmt.Println(Summary(totalIssues, len(distinctRules), len(fileOrder)))
	return nil
}
