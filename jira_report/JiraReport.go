package main

import (
	"encoding/json"
	"fmt"
	"jira_report/search"
	"jira_report/users"
	"jira_report/util"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/blinkbean/dingtalk"
	"github.com/shopspring/decimal"
)

type Assignee struct {
	DisplayName string `json:"displayName"`
}

type Fields struct {
	Assignee Assignee
}

type ISSItem struct {
	Fields Fields
}

type JiraInfo struct {
	Expand     string    `json:"expand"`
	StartAt    int       `json:"startAt"`
	MaxResults int       `json:"maxResults"`
	Total      int       `json:"total"`
	Issues     []ISSItem `json:"issues"`
}

func removeDuplication_map(arr []string) []string {
	set := make(map[string]struct{}, len(arr))
	j := 0
	for _, v := range arr {
		_, ok := set[v]
		if ok {
			continue
		}
		set[v] = struct{}{}
		arr[j] = v
		j++
	}
	return arr[:j]
}

func GetCompleteTotal(JiraURL string) int {
	yesterday, nowtime := util.GetTime()
	JiraSourceURL := JiraURL + "/rest/api/2/search"
	JiraSQL := "status in (Resolved, Closed, 完成) AND updated >= " + yesterday + " AND updated < " + nowtime
	JiraRep := search.RequestJira(JiraSourceURL, JiraSQL, "0")
	var jirainfo JiraInfo
	err := json.Unmarshal([]byte(JiraRep), &jirainfo)
	if err != nil {
		fmt.Println(err)
	}
	return jirainfo.Total
}

func GetNewJiraTotal(JiraURL string) int {
	yesterday, nowtime := util.GetTime()
	JiraSourceURL := JiraURL + "/rest/api/2/search"
	JiraSQL := "status in (待办, 待处理) AND created >= " + yesterday + " AND created < " + nowtime
	JiraRep := search.RequestJira(JiraSourceURL, JiraSQL, "0")
	var jirainfo JiraInfo
	err := json.Unmarshal([]byte(JiraRep), &jirainfo)
	if err != nil {
		fmt.Println(err)
	}
	return jirainfo.Total
}

func GetTimeoutJira(JiraURL string) (int, []string) {
	JiraSourceURL := JiraURL + "/rest/api/2/search"
	JiraSQL := "issuetype not in (沟通支持, 沟通子任务, Epic) AND resolution = Unresolved AND due < '0'"
	JiraRep := search.RequestJira(JiraSourceURL, JiraSQL, "0")
	var jirainfo JiraInfo
	var jirausers []string
	err := json.Unmarshal([]byte(JiraRep), &jirainfo)
	if err != nil {
		fmt.Println(err)
	}
	for _, issue := range jirainfo.Issues {
		jirausers = append(jirausers, issue.Fields.Assignee.DisplayName)
	}
	return jirainfo.Total, removeDuplication_map(jirausers)
}

func main() {
	JiraToken := os.Getenv("jira_token")
	JiraURL := os.Getenv("jira_url")
	yesterday, _ := util.GetTime()
	timeoutTotal, timeoutUsers := GetTimeoutJira(JiraURL)
	NewJira := GetNewJiraTotal(JiraURL)
	CompleteTotal := GetCompleteTotal(JiraURL)
	var workTimeNoWrite []string
	WorkTimeTotal := 0
	for _, value := range users.GetJiraUser(JiraURL) {
		UserJiraWorktime := users.CheckJiraTime(value.Key, JiraURL)
		WorkTimeTotal += UserJiraWorktime
		if UserJiraWorktime < 14400 {
			workTimeNoWrite = append(workTimeNoWrite, value.DisplayName)
		}
	}
	resultTime, _ := decimal.NewFromFloat(float64(WorkTimeTotal) / 3600).Round(2).Float64()
	// https://oapi.dingtalk.com/robot/send?access_token=7387469a96fa2bc3e43ccc075d6ce658c63a70b994af0d0412198a51dc69f26d
	var dingToken = []string{JiraToken}
	dingClient := dingtalk.InitDingTalk(dingToken, "日报")
	dingContent := "**研发任务执行情况日报** \n\n" +
		"**统计日期**: " + string(yesterday) + " \n\n" +
		"**新建任务数**: " + strconv.Itoa(NewJira) + " \n\n" +
		"**完成任务数**: " + strconv.Itoa(CompleteTotal) + " \n\n" +
		"**任务超时未完成数**: " + strconv.Itoa(timeoutTotal) + " \n\n" +
		"**超时任务查询地址**: " + JiraURL + "/issues/?filter=12110 \n\n" +
		"**任务超时未完成人员**: " + strings.Join(timeoutUsers, " ") + " \n\n" +
		"**完成研发工时**: " + fmt.Sprintf("%v", resultTime) + "h \n\n" +
		"**未按时填写工时人数**: " + strconv.Itoa(len(workTimeNoWrite)) + " \n\n" +
		"**未按时填写工时人员**: " + strings.Join(workTimeNoWrite, " ") + " \n\n"
	err := dingClient.SendMarkDownMessage("Jira日报助手", dingContent)
	if err != nil {
		log.Println(err)
	}
}
