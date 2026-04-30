package doctor

import (
	"os"
	"path/filepath"
	"strings"
)

type Status int

const (
	OK Status = iota
	Warn
	Fail
)

type Check struct {
	Name    string
	Status  Status
	Message string
}

type Result struct {
	Checks []Check
}

func (status Status) String() string {
	switch status {
	case Warn:
		return "warn"
	case Fail:
		return "fail"
	default:
		return "ok"
	}
}

func (result Result) HasFailures() bool {
	for _, check := range result.Checks {
		if check.Status == Fail {
			return true
		}
	}
	return false
}

func Run(root string) Result {
	var result Result
	result.add(pathExists(root, ".knowblazer/config.json", false, "config"))
	for _, dir := range requiredDirs {
		result.add(pathExists(root, dir, true, dir))
	}
	for _, file := range requiredFiles {
		result.add(pathExists(root, file, false, file))
	}
	result.add(quarantineCheck(root))
	result.add(gitRepoCheck(root))
	result.add(gitRemoteCheck(root))
	return result
}

var requiredDirs = []string{
	"inbox",
	"daily",
	"projects",
	"experience",
	"experience/auto",
	"quarantine",
}

var requiredFiles = []string{
	"AI-SETUP.md",
}

func (result *Result) add(check Check) {
	result.Checks = append(result.Checks, check)
}

func pathExists(root string, rel string, wantDir bool, name string) Check {
	path := filepath.Join(root, rel)
	info, err := os.Stat(path)
	if err != nil {
		return Check{Name: name, Status: Fail, Message: "missing " + rel}
	}
	if wantDir && !info.IsDir() {
		return Check{Name: name, Status: Fail, Message: rel + " is not a directory"}
	}
	if !wantDir && info.IsDir() {
		return Check{Name: name, Status: Fail, Message: rel + " is not a file"}
	}
	return Check{Name: name, Status: OK, Message: rel + " exists"}
}

func quarantineCheck(root string) Check {
	quarantineRoot := filepath.Join(root, "quarantine")
	if _, err := os.Stat(quarantineRoot); err != nil {
		return Check{Name: "quarantine", Status: Fail, Message: "quarantine is missing"}
	}
	count := 0
	err := filepath.WalkDir(quarantineRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() == "README.md" {
			return nil
		}
		count++
		return nil
	})
	if err != nil {
		return Check{Name: "quarantine-files", Status: Fail, Message: "could not inspect quarantine"}
	}
	if count > 0 {
		return Check{Name: "quarantine-files", Status: Warn, Message: "quarantine contains files to review"}
	}
	return Check{Name: "quarantine-files", Status: OK, Message: "no quarantine files"}
}

func gitRepoCheck(root string) Check {
	info, err := os.Stat(filepath.Join(root, ".git"))
	if err != nil {
		return Check{Name: "git-repo", Status: Warn, Message: "not a Git repo"}
	}
	if !info.IsDir() {
		return Check{Name: "git-repo", Status: Warn, Message: ".git is not a directory"}
	}
	return Check{Name: "git-repo", Status: OK, Message: "Git repo detected"}
}

func gitRemoteCheck(root string) Check {
	content, err := os.ReadFile(filepath.Join(root, ".git", "config"))
	if err != nil {
		return Check{Name: "git-remote", Status: Warn, Message: "no Git remote configured"}
	}
	if strings.Contains(string(content), "[remote ") {
		return Check{Name: "git-remote", Status: OK, Message: "Git remote configured"}
	}
	return Check{Name: "git-remote", Status: Warn, Message: "no Git remote configured"}
}
