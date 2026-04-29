package scan

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Level int

const (
	Clean Level = iota
	Warning
	High
)

type Finding struct {
	File    string
	Line    int
	Rule    string
	Level   Level
	Snippet string
}

type Result struct {
	Level    Level
	Findings []Finding
}

func (level Level) String() string {
	switch level {
	case High:
		return "high"
	case Warning:
		return "warning"
	default:
		return "clean"
	}
}

type rule struct {
	name    string
	level   Level
	pattern *regexp.Regexp
	redact  func(string) string
}

var highRules = []rule{
	{
		name:    "private-key",
		level:   High,
		pattern: regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`),
		redact:  redactWholeLine,
	},
	{
		name:    "database-url",
		level:   High,
		pattern: regexp.MustCompile(`(?i)(mysql|postgres|postgresql|mongodb)://[^\s]+`),
		redact:  redactDatabaseURL,
	},
	{
		name:    "secret-field",
		level:   High,
		pattern: regexp.MustCompile(`(?i)\b(token|api_key|apikey|secret|password)\s*[:=]\s*['"]?[^'"\s]+`),
		redact:  redactAssignment,
	},
	{
		name:    "cloud-secret-env",
		level:   High,
		pattern: regexp.MustCompile(`(?i)\b(AWS_SECRET_ACCESS_KEY|GOOGLE_APPLICATION_CREDENTIALS|AZURE_CLIENT_SECRET)\s*=\s*[^\s]+`),
		redact:  redactAssignment,
	},
	{
		name:    "bearer-token",
		level:   High,
		pattern: regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+/=-]{12,}`),
		redact:  redactBearer,
	},
	{
		name:    "openai-style-key",
		level:   High,
		pattern: regexp.MustCompile(`\bsk-[A-Za-z0-9_-]{12,}`),
		redact:  redactOpenAIStyleKey,
	},
}

func Path(path string) (Result, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Result{}, err
	}
	if info.IsDir() {
		return scanDir(path)
	}
	return scanFile(path)
}

func scanDir(root string) (Result, error) {
	var combined Result
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if shouldSkipDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		result, err := scanFile(path)
		if err != nil {
			return err
		}
		combined.Findings = append(combined.Findings, result.Findings...)
		if result.Level > combined.Level {
			combined.Level = result.Level
		}
		return nil
	})
	return combined, err
}

func shouldSkipDir(name string) bool {
	switch name {
	case ".git", "node_modules", ".idea", ".vscode":
		return true
	default:
		return false
	}
}

func scanFile(path string) (Result, error) {
	file, err := os.Open(path)
	if err != nil {
		return Result{}, err
	}
	defer file.Close()

	var result Result
	scanner := bufio.NewScanner(file)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		for _, rule := range highRules {
			if rule.pattern.MatchString(line) {
				finding := Finding{
					File:    path,
					Line:    lineNo,
					Rule:    rule.name,
					Level:   rule.level,
					Snippet: rule.redact(line),
				}
				result.Findings = append(result.Findings, finding)
				if rule.level > result.Level {
					result.Level = rule.level
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return Result{}, err
	}
	return result, nil
}

func redactWholeLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return "****"
	}
	return "****"
}

func redactAssignment(line string) string {
	re := regexp.MustCompile(`(?i)^(\s*[^:=\s]+\s*[:=]\s*)['"]?([^'"\s]+)(.*)$`)
	return re.ReplaceAllString(line, "${1}****${3}")
}

func redactDatabaseURL(line string) string {
	re := regexp.MustCompile(`(?i)((mysql|postgres|postgresql|mongodb)://[^:\s]+:)[^@\s]+(@[^\s]+)`)
	if re.MatchString(line) {
		return re.ReplaceAllString(line, "${1}****${3}")
	}
	re = regexp.MustCompile(`(?i)(mysql|postgres|postgresql|mongodb)://[^\s]+`)
	return re.ReplaceAllStringFunc(line, func(match string) string {
		parts := strings.SplitN(match, "://", 2)
		if len(parts) != 2 {
			return "****"
		}
		return fmt.Sprintf("%s://****", parts[0])
	})
}

func redactBearer(line string) string {
	re := regexp.MustCompile(`(?i)(\bBearer\s+)[A-Za-z0-9._~+/=-]{12,}`)
	return re.ReplaceAllString(line, "${1}****")
}

func redactOpenAIStyleKey(line string) string {
	re := regexp.MustCompile(`\bsk-[A-Za-z0-9_-]{12,}`)
	return re.ReplaceAllString(line, "sk-****")
}
