package search

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

func RequestJira(jiraURL string, jiraSQL string, startAt string) string {
	JiraUser := os.Getenv("jira_user")
	JiraPassword := os.Getenv("jira_password")
	client := http.Client{Timeout: 5 * time.Second}
	q := url.Values{}
	q.Add("startAt", startAt)
	q.Add("maxResults", "200")
	if jiraSQL == "" {
		q.Add("groupname", "lckj-dev")
	} else {
		q.Add("jql", jiraSQL)
	}
	req, err := http.NewRequest("GET", jiraURL, nil)
	if err != nil {
		panic(err)
	}
	req.URL.RawQuery = q.Encode()
	req.SetBasicAuth(JiraUser, JiraPassword)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	return string(resBody)
}
