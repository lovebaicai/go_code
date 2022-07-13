package users

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"jira_report/search"
	"jira_report/util"
	"log"
	"net/http"
	"os"
	"time"
)

type Value struct {
	Key         string `json:"key"`
	DisplayName string `json:"displayName"`
}

type UserInfo struct {
	Value []Value `json:"values"`
}

type UserBody struct {
	From   string   `json:"from"`
	To     string   `json:"to"`
	Worker []string `json:"worker"`
}

type JiraTime struct {
	BillableSeconds int `json:"billableSeconds"`
}

func GetJiraUser(JiraURL string) []Value {
	JiraSourceURL := JiraURL + "/rest/api/2/group/member"
	JiraRep := search.RequestJira(JiraSourceURL, "", "0")
	JiraRep2 := search.RequestJira(JiraSourceURL, "", "50")
	var userinfo UserInfo
	var userinfo2 UserInfo
	err := json.Unmarshal([]byte(JiraRep), &userinfo)
	if err != nil {
		fmt.Println(err)
	}
	err = json.Unmarshal([]byte(JiraRep2), &userinfo2)
	if err != nil {
		fmt.Println(err)
	}
	allUser := append(userinfo.Value, userinfo2.Value...)
	return allUser
}

func CheckJiraTime(CheckUser string, JiraURL string) int {
	JiraUser := os.Getenv("jira_user")
	JiraPassword := os.Getenv("jira_password")
	_, yesterday := util.GetTime()
	var CheckArray []string
	CheckArray = append(CheckArray, CheckUser)
	userInfo := UserBody{
		From:   yesterday,
		To:     yesterday,
		Worker: CheckArray,
	}
	body, _ := json.Marshal(userInfo)
	req, err := http.NewRequest("POST", JiraURL+"/rest/tempo-timesheets/4/worklogs/search", bytes.NewBuffer(body))
	if err != nil {
		log.Println(nil)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(JiraUser, JiraPassword)
	client := http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	var jiratime []JiraTime
	err = json.Unmarshal([]byte(string(resBody)), &jiratime)
	if err != nil {
		panic(err)
	}
	timeSum := 0
	for _, value := range jiratime {
		timeSum += value.BillableSeconds
	}
	return timeSum
}
