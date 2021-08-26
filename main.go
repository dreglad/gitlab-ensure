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

var glClient *gitlab.Client

func init() {
	godotenv.Load()

	url := os.Getenv("GITLAB_URL")
	token := os.Getenv("GITLAB_TOKEN")
	if token == "" {
		fmt.Printf("Token required in env as GITLAB_TOKEN")
		os.Exit(1)
	}

	git, err := gitlab.NewClient(token, gitlab.WithBaseURL(url))
	if err != nil {
		fmt.Printf("error creating GitLab client: %v", err)
		os.Exit(1)
	}

	glClient = git
}

func main() {
	ch := make(chan *gitlab.Project)
	var wg sync.WaitGroup
	for _, pid := range strings.Split(os.Getenv("KNOWN_OPEN"), ",") {
		wg.Add(1)
		go ensureProject(
			strings.Trim(pid, " "),
			gitlab.Visibility("public"),
			ch, &wg)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for project := range ch {
		fmt.Println("Project", project.Name)
	}
}

func ensureProject(pid string, v *gitlab.VisibilityValue, ch chan *gitlab.Project, wg *sync.WaitGroup) error {
	defer (*wg).Done()

	getOpt := &gitlab.GetProjectOptions{}
	p, _, err := glClient.Projects.GetProject(pid, getOpt)
	if err != nil {
		return fmt.Errorf("error getting project: %v", err)
	}

	if p.Visibility != *v {
		editOpt := &gitlab.EditProjectOptions{Visibility: v}
		p, _, err = glClient.Projects.EditProject(p.ID, editOpt)
		if err != nil {
			return fmt.Errorf("error editing project: %v", err)
		}
		ch <- p
	}

	return nil
}
