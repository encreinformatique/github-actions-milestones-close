package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Milestone struct {
	Title     string `json:"title"`
	DueOn     string `json:"due_on"`
	URL       string `json:"url"`
	Number    int    `json:"number"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
}

type config struct {
	Token      string `envconfig:"GITHUB_TOKEN" required:"true"`
	Owner      string `envconfig:"GITHUB_REPOSITORY_OWNER" required:"true"`
	Repository string `envconfig:"GITHUB_REPOSITORY" required:"true"`
}

type Issue struct {
	URL string `json:"url"`
}

func main() {
	var c config
	if c.Token == "" {
		fmt.Println("Error: GitHub token not provided")
		os.Exit(1)
	}

	if c.Owner == "" {
		fmt.Println("Error: GitHub repository owner not provided")
		os.Exit(1)
	}

	if c.Repository == "" {
		fmt.Println("Error: GitHub repository not provided")
		os.Exit(1)
	}

	daysPastDue := 7 // Number of days past-due to close
	now := time.Now()

	// Get all open milestones
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/milestones?state=open", c.Owner, c.Repository)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
	request.Header.Add("Authorization", fmt.Sprintf("token %s", c.Token))
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
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?state=open&milestone=%d", c.Owner, c.Repository, milestone.Number)
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("Error: %v", err)
			os.Exit(1)
		}
		request.Header.Add("Authorization", fmt.Sprintf("token %s", c.Token))
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
