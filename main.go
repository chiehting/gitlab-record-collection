package main

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/chiehting/gitlab-record-collection/pkg/config"
	"github.com/chiehting/gitlab-record-collection/pkg/log"
	"github.com/chiehting/gitlab-record-collection/services"
	"github.com/go-co-op/gocron"
)

func init() {
	services.Service.Init()
}

func main() {
	awsClient := services.AWS
	awsClient.CreateSession()
	awsClient.ConnectionCloudwatchLog()
	gitlab := services.Gitlab

	// 要傳送的日誌訊息
	s := gocron.NewScheduler(time.UTC)
	dateFormat := "200601"

	GitLabs := config.GetGitLabs()
	for _, item := range *GitLabs {
		GitLab := item

		now := time.Now()
		awsClient.CreateLogStream(GitLab.Domain, "project"+"-"+now.Format(dateFormat))
		awsClient.CreateLogStream(GitLab.Domain, "commit"+"-"+now.Format(dateFormat))

		s.Every(1).MonthLastDay().Do(func() {
			currentTime := time.Now()
			twoDayAgo := currentTime.Add(48 * time.Hour)
			awsClient.CreateLogStream(GitLab.Domain, "project"+"-"+twoDayAgo.Format(dateFormat))
			awsClient.CreateLogStream(GitLab.Domain, "commit"+"-"+twoDayAgo.Format(dateFormat))
		})

		// 傳送專案內容
		_, err := s.Every(1).Day().Do(func() {
			currentTime := time.Now()
			currentFormat := currentTime.Format(dateFormat)

			projects := gitlab.GetProjects(GitLab)
			projectLogEvents := []*cloudwatchlogs.InputLogEvent{}

			for _, project := range projects {
				logData := map[string]interface{}{
					"domain":              GitLab.Domain,
					"id":                  project.ID,
					"name":                project.Name,
					"path":                project.Path,
					"path_with_namespace": project.PathWithNamespace,
					"last_activity_at":    project.LastActivityAt,
				}

				logJSON, err := json.Marshal(logData)
				if err != nil {
					log.Error("無法轉換成 JSON：", err)
					continue
				}

				logEvent := &cloudwatchlogs.InputLogEvent{
					Message:   aws.String(string(logJSON)),
					Timestamp: aws.Int64(aws.TimeUnixMilli(time.Now())),
				}

				projectLogEvents = append(projectLogEvents, logEvent)

				if len(projectLogEvents) == 100 {
					log.Debug(projectLogEvents)
					awsClient.PutLogEvents(GitLab.Domain, "project"+"-"+currentFormat, projectLogEvents)
					projectLogEvents = []*cloudwatchlogs.InputLogEvent{}
				}
			}

			if len(projectLogEvents) != 0 {
				log.Debug(projectLogEvents)
				awsClient.PutLogEvents(GitLab.Domain, "project"+"-"+currentFormat, projectLogEvents)
			}
		})
		if err != nil {
			log.Error("Get projects", err)
		}

		// 傳送提交內容
		_, err = s.Every(1).Hour().Do(func() {
			currentTime := time.Now()
			currentFormat := currentTime.Format(dateFormat)

			commitsWithProject := gitlab.GetCommits(GitLab)
			commitLogEvents := []*cloudwatchlogs.InputLogEvent{}

			// 使用迴圈處理 JSON 資料
			for _, commitWithProject := range commitsWithProject {
				logData := map[string]interface{}{
					"domain":         GitLab.Domain,
					"project_id":     commitWithProject.ProjectID,
					"project_name":   commitWithProject.ProjectName,
					"id":             commitWithProject.Commit.ShortID,
					"title":          commitWithProject.Commit.Title,
					"author_name":    commitWithProject.Commit.AuthorName,
					"author_email":   commitWithProject.Commit.AuthorEmail,
					"committed_date": commitWithProject.Commit.CommittedDate,
				}

				logJSON, err := json.Marshal(logData)
				if err != nil {
					log.Error("無法轉換成 JSON：", err)
					continue
				}

				parsedTime, err := time.Parse(time.RFC3339, commitWithProject.Commit.CommittedDate)
				if err != nil {
					log.Error("無法解析時間字串：", err)
					return
				}

				logEvent := &cloudwatchlogs.InputLogEvent{
					Message:   aws.String(string(logJSON)),
					Timestamp: aws.Int64(aws.TimeUnixMilli(parsedTime)),
				}

				commitLogEvents = append(commitLogEvents, logEvent)
				if len(commitLogEvents) == 100 {
					log.Debug(commitLogEvents)
					awsClient.PutLogEvents(GitLab.Domain, "commit"+"-"+currentFormat, commitLogEvents)
					commitLogEvents = []*cloudwatchlogs.InputLogEvent{}
				}

			}

			if len(commitLogEvents) != 0 {
				log.Debug(commitLogEvents)
				awsClient.PutLogEvents(GitLab.Domain, "commit"+"-"+currentFormat, commitLogEvents)
			}
		})
		if err != nil {
			log.Error("Get commits", err)
		}
	}

	s.StartAsync()
	s.StartBlocking()
}
