package utils

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func SanitizeBranchArray(branches []string) []string {
	var sanitizedBranches []string
	for _, branch := range branches {
		trimmedBranch := strings.TrimSpace(branch)

		if trimmedBranch != "" {
			if strings.HasPrefix(branch, "*") {
				defaultBranchTrimmed := strings.TrimPrefix(branch, "* ")
				sanitizedBranches = append(sanitizedBranches, defaultBranchTrimmed)
			} else {
				sanitizedBranches = append(sanitizedBranches, trimmedBranch)
			}
		}
	}
	return sanitizedBranches
}

func FetchLocalBranches(repoPath *string) []string {
	allBranchesList := exec.Command("git", "branch", "--list")
	allBranchesList.Dir = *repoPath

	output, err := allBranchesList.Output()
	if err != nil {
		fmt.Println("Error listing local branches:", err)
		os.Exit(1)
	}

	branches := strings.Split(string(output), "\n")
	sanitizedBranches := SanitizeBranchArray(branches)
	return sanitizedBranches
}

func IsBranchAlreadyMerged(branch string, mergedIntoMainBranches []string) bool {
	for _, branchMergedIntoMain := range mergedIntoMainBranches {
		if strings.EqualFold(branchMergedIntoMain, branch) {
			return true
		}
	}

	return false
}

func GetCurrentBranch(repoPath *string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = *repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error getting current branch: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func ShortenBranchName(branch string) string {
	const maxBranchNameLength = 25
	if len(branch) > maxBranchNameLength {
		return branch[:maxBranchNameLength-3] + "..."
	}
	return branch
}

func AskForConfirmation(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt + " ")
	response, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			fmt.Println("Exiting the program.")
			os.Exit(0)
		} else {
			fmt.Println("Error reading input:", err)
			os.Exit(1)
		}
	}
	response = strings.TrimSpace(response)
	return strings.EqualFold(response, "y") || strings.EqualFold(response, "yes")
}

func IsGitRepository() bool {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current working directory:", err)
		return false
	}

	gitPath := filepath.Join(cwd, ".git")
	_, err = os.Stat(gitPath)
	if os.IsNotExist(err) {
		return false
	} else if err != nil {
		fmt.Println("Error checking for .git directory:", err)
		return false
	}

	return true
}
