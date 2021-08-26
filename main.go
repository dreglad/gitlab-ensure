// Ensures projects have the correct visibility.
package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/xanzy/go-gitlab"
)

var (
	glClient *gitlab.Client
	wg       sync.WaitGroup
)

func init() {
	godotenv.Load()

	url := os.Getenv("GITLAB_URL")
	token := os.Getenv("GITLAB_TOKEN")

	if token == "" {
		handleError(fmt.Errorf("GITLAB_TOKEN is not set"))
	}

	git, err := gitlab.NewClient(token, gitlab.WithBaseURL(url))
	if err != nil {
		handleError(fmt.Errorf("creating GitLab client: %v", err))
	}

	glClient = git
}

func main() {
	ch := make(chan *gitlab.Project)
	for _, pid := range splitCommas(os.Getenv("KNOWN_OPEN")) {
		wg.Add(1)
		go ensureProject(pid, "public", ch)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for project := range ch {
		fmt.Printf("Updated project %v\n", project.Name)
	}
}

// Ensure project has the desired visibility.
func ensureProject(pid string, v gitlab.VisibilityValue, ch chan *gitlab.Project) error {
	defer wg.Done()

	getOpt := &gitlab.GetProjectOptions{}
	proj, _, err := glClient.Projects.GetProject(pid, getOpt)
	if err != nil {
		return fmt.Errorf("when getting project: %v", err)
	}

	if proj.Visibility != v {
		editOpt := &gitlab.EditProjectOptions{Visibility: gitlab.Visibility(v)}
		proj, _, err = glClient.Projects.EditProject(proj.ID, editOpt)
		if err != nil {
			return fmt.Errorf("when editing project: %v", err)
		}
		ch <- proj
	}

	return nil
}

// Split string on commas and trim any whitespace.
func splitCommas(s string) []string {
	return strings.Split(strings.ReplaceAll(s, " ", ""), ",")
}

// Print error to stoud and exit with non-zero exit code.
func handleError(err error) {
	fmt.Print(err)
	os.Exit(1)
}
