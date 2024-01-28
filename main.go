package main

import (
	"fmt"
	"gcyb/utils"
	"os"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/eiannone/keyboard"
	"github.com/spf13/cobra"
)

var defaultRepoPath string = "."
var repoPath *string = &defaultRepoPath
var branchAndReasonSeparator string = " | "

var defaultBranches = map[string]bool{
	"master":      true,
	"main":        true,
	"development": true,
	"dev":         true,
	"testing":     true,
	"test":        true,
}

var rootCmd = &cobra.Command{
	Use:   "gcyb",
	Short: "Displays a table of branches that could be cleaned and the reason of it. It does not delete branches.",
	Long: "gcyb (Go Clean Your Branches) is a CLI tool for Git to detect and, optionally, delete branches " +
		"that were already merged in your current branch/HEAD. This CLI tool will never delete or update " +
		"remote branches without your permission. The commands of reading and deleting branches are separated " +
		"to avoid possible unwanted cleaning of programmers (branches).",
	Run: runGcybDryReadCommand,
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Deletes the deletable branches after 2 confirmations.",
	Long:  "Delete every branch that is considered 'deletable'. For a branch to be elegible as 'deletable', it needs to be already merged in your branch. Stale branches, meaning branches that are not receiving commits for a long time, should be analyzed by the user and not by a third party app due to not knowing if the commits are safe to lose.",
	Run:   runDeleteBranchesCommand,
}

var pickCmd = &cobra.Command{
	Use:   "pick",
	Short: "Select which branches you want to delete.",
	Run:   runDeleteBranchesCommand,
}

func init() {
	if !utils.IsGitRepository() {
		fmt.Println("Not a git repository.")
		os.Exit(1)
	}

	rootCmd.AddCommand(cleanCmd, pickCmd)
	rootCmd.PersistentFlags().StringVarP(repoPath, "repo", "r", defaultRepoPath, "Specify a local path to a git repository.")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func Checkboxes(label string, opts []string) []string {
	res := []string{}
	prompt := &survey.MultiSelect{
		Message: label,
		Options: opts,
	}
	err := survey.AskOne(prompt, &res, survey.WithRemoveSelectAll(), survey.WithRemoveSelectNone())
	if err != nil {
		if err == terminal.InterruptErr {
			fmt.Println("Interrupted")
			os.Exit(1)
		}
	}

	var selectedBranches []string
	for _, selected := range res {
		branchName := strings.Split(selected, branchAndReasonSeparator)
		if len(branchName) > 0 {
			selectedBranches = append(selectedBranches, branchName[0])
		}
	}

	return selectedBranches
}

func runGcybDryReadCommand(cmd *cobra.Command, args []string) {
	branches := utils.FetchLocalBranches(repoPath)
	currentBranch, err := utils.GetCurrentBranch(repoPath)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	deletableBranches, reasonOfDeletion := checkDeletableBranches(branches, repoPath)

	if len(deletableBranches) == 0 {
		fmt.Print("Nothing to clean!")
		os.Exit(0)
	}

	displayDeletableBranchesTable(currentBranch, deletableBranches, reasonOfDeletion)
}

func runDeleteBranchesCommand(cmd *cobra.Command, args []string) {
	branches := utils.FetchLocalBranches(repoPath)
	deletableBranches, reasonOfDeletion := checkDeletableBranches(branches, repoPath)

	if len(deletableBranches) == 0 {
		fmt.Print("Nothing to clean!")
		os.Exit(0)
	}

	currentBranch, err := utils.GetCurrentBranch(repoPath)

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	reasons := make(map[string]string)

	if cmd.Use == "pick" {
		err := keyboard.Open()
		if err != nil {
			fmt.Println(err)
		}
		defer func() {
			_ = keyboard.Close()
		}()

		for _, branch := range deletableBranches {
			reasons[branch] = reasonOfDeletion
		}

		selectedBranches := Checkboxes(
			"Select branches to delete",
			appendReasonToDeletableBranch(deletableBranches, reasons),
		)

		if len(selectedBranches) == 0 {
			fmt.Print("No branches selected")
			os.Exit(0)
		}

		fmt.Print(selectedBranches)
		waitForConfirmationToDeleteBranches(selectedBranches)
	}

	if cmd.Use == "clean" {
		displayDeletableBranchesTable(currentBranch, deletableBranches, reasonOfDeletion)
		waitForConfirmationToDeleteBranches(deletableBranches)
	}
}

func waitForConfirmationToDeleteBranches(deletableBranches []string) {
	userResponse := utils.AskForConfirmation("Do you want to proceed and delete these branches? (y/n)")
	if userResponse {
		additionalConfirmation := utils.AskForConfirmation("Just to be sure, you are about to delete " + fmt.Sprint(len(deletableBranches)) + " branches. Confirm? (y/n)")
		if additionalConfirmation {
			deleteBranches(deletableBranches)
		}
	} else {
		fmt.Println("No branches were deleted.")
	}
}

func checkDeletableBranches(branches []string, repoPath *string) ([]string, string) {
	var deletableBranches []string
	alreadyMergedBranch := exec.Command("git", "branch", "--merged")
	alreadyMergedBranch.Dir = *repoPath
	alreadyMergedBranchesOutput, err := alreadyMergedBranch.Output()
	if err != nil {
		fmt.Print(err)
	}

	currentBranch, err := utils.GetCurrentBranch(repoPath)
	if err != nil {
		fmt.Println("Error:", err)
	}

	alreadyMergedBranchesList := strings.Split(string(alreadyMergedBranchesOutput), "\n")

	alreadyMergedBranchesList = utils.SanitizeBranchArray(alreadyMergedBranchesList)

	var reasonOfDeletion string
	for _, branch := range branches {

		var isDefaultOrCurrentBranch bool = defaultBranches[branch] || strings.EqualFold(branch, currentBranch)
		if isDefaultOrCurrentBranch {
			continue
		}

		if !utils.IsBranchAlreadyMerged(branch, alreadyMergedBranchesList) {
			continue
		} else {
			reasonOfDeletion = "Already merged in your current branch."
		}

		deletableBranches = append(deletableBranches, branch)
	}

	return deletableBranches, reasonOfDeletion
}

func appendReasonToDeletableBranch(branches []string, reasons map[string]string) []string {
	var branchList []string
	for _, branch := range branches {
		reason, exists := reasons[branch]
		if exists {
			branchList = append(branchList, strings.TrimSpace(branch)+branchAndReasonSeparator+reason)
		} else {
			branchList = append(branchList, branch)
		}
	}
	return branchList
}

func displayDeletableBranchesTable(currentBranch string, deletableBranches []string, reasonOfDeletion string) {

	defaultStyle := lipgloss.NewStyle()
	reasonStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#d4db09"))
	currentBranchStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#49c706"))

	formattedRows := make([][]string, 0)
	for _, deletableBranch := range deletableBranches {
		formattedRows = append(formattedRows, []string{deletableBranch, reasonOfDeletion})
	}

	t := table.New().
		Width(59).
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().
			Foreground(lipgloss.Color("0xFFFF"))).
		StyleFunc(func(row int, col int) lipgloss.Style {
			if col == 1 && row != 0 {
				return reasonStyle
			}
			return defaultStyle
		}).
		Headers("Branch", "Reason for deletion").
		Rows(formattedRows...)

	fmt.Println(t)
	fmt.Println("Current branch: " + currentBranchStyle.Render(currentBranch))
}

func deleteBranches(branches []string) {
	for _, branch := range branches {
		deleteCmd := exec.Command("git", "branch", "-d", branch)
		deleteCmd.Dir = *repoPath
		output, err := deleteCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error deleting branch '%s' %v\n", branch, err)
			fmt.Printf("Command output: %s", output)
			os.Exit(1)
		} else {
			fmt.Printf("Branch %s deleted successfully.\n", branch)
		}
	}

	fmt.Printf("Cleaned a total of %d branches", len(branches))
	os.Exit(0)
}
