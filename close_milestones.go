package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Milestone struct {
	Title     string `json:"title"`
	DueOn     string `json:"due_on"`
	URL       string `json:"url"`
	Number    int    `json:"number"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
}

type Issue struct {
	URL string `json:"url"`
}

func main() {
	token      string `envconfig:"GITHUB_TOKEN" required:"true"`
	if token == "" {
		fmt.Println("Error: GitHub token not provided")
		os.Exit(1)
	}

	owner      string `envconfig:"GITHUB_REPOSITORY_OWNER" required:"true"`
	if owner == "" {
		fmt.Println("Error: GitHub repository owner not provided")
		os.Exit(1)
	}

	repo      string `envconfig:"GITHUB_REPOSITORY" required:"true"`
	if repo == "" {
		fmt.Println("Error: GitHub repository not provided")
		os.Exit(1)
	}

	daysPastDue := 7 // Number of days past-due to close
	now := time.Now()

	// Get all open milestones
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/milestones?state=open", owner, repo)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
	request.Header.Add("Authorization", fmt.Sprintf("token %s", token))
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
	defer response.Body.Close()

	var milestones []Milestone
	err = json.NewDecoder(response.Body).Decode(&milestones)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	// Loop through each milestone and check if it's past-due
	for _, milestone := range milestones {
		if milestone.State != "open" || milestone.DueOn == "" {
			continue
		}

		dueOn, err := time.Parse(time.RFC3339, milestone.DueOn)
		if err != nil {
			fmt.Printf("Error: %v", err)
			os.Exit(1)
		}

		daysPast := int(now.Sub(dueOn).Hours() / 24)
		if daysPast < daysPastDue {
			continue
		}

		// Check if the milestone has any open issues
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?state=open&milestone=%d", owner, repo, milestone.Number)
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("Error: %v", err)
			os.Exit(1)
		}
		request.Header.Add("Authorization", fmt.Sprintf("token %s", token))
		client := &http.Client{}
		response, err := client.Do(request)
		if err != nil {
			fmt.Printf("Error: %v", err)
			os.Exit(1)
		}
		defer response.Body.Close()

		var issues []Issue
		err = json.NewDecoder(response.Body).Decode(&issues)
		if err != nil {
			fmt.Printf("Error: %v", err)
			os.Exit(1)
		}

		if len(issues) > 0 {
			fmt.Printf("Milestone '%s' still has open issues\n", milestone)
		}
	}
}
