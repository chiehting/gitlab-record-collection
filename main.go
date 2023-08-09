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

	GitLabs := config.GetGitLabs()
	for _, item := range *GitLabs {
		GitLab := item

		now := time.Now()
		todayFormat := ""
		awsClient.CreateLogStream(GitLab.Domain, "project"+"-"+now.Format("20060102"))
		awsClient.CreateLogStream(GitLab.Domain, "commit"+"-"+now.Format("20060102"))

		_, err := s.Every(1).Day().Do(func() {
			currentTime := time.Now()
			tomorrowTIme := currentTime.Add(24 * time.Hour)
			todayFormat = currentTime.Format("20060102")
			tomorrowFormat := tomorrowTIme.Format("20060102")
			awsClient.CreateLogStream(GitLab.Domain, "project"+"-"+tomorrowFormat)
			awsClient.CreateLogStream(GitLab.Domain, "commit"+"-"+tomorrowFormat)
		})
		if err != nil {
			log.Error("Create Log Stream", err)
		}

		// 傳送專案
		_, err = s.Every(1).Day().Do(func() {
			projects := gitlab.GetProjects(GitLab)
			logEvents := []*cloudwatchlogs.InputLogEvent{}

			// 使用迴圈處理 JSON 資料
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

				logEvents = append(logEvents, logEvent)
			}

			if len(logEvents) != 0 {
				log.Debug(logEvents)
				awsClient.PutLogEvents(GitLab.Domain, "project"+"-"+todayFormat, logEvents)
			}
		})
		if err != nil {
			log.Error("Get projects", err)
		}

		_, err = s.Every(1).Hour().Do(func() {
			commitsWithProject := gitlab.GetCommits(GitLab)
			logEvents2 := []*cloudwatchlogs.InputLogEvent{}

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
					"message":        commitWithProject.Commit.Message,
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

				logEvents2 = append(logEvents2, logEvent)
			}

			if len(logEvents2) != 0 {
				log.Debug(logEvents2)
				awsClient.PutLogEvents(GitLab.Domain, "commit"+"-"+todayFormat, logEvents2)
			}
		})
		if err != nil {
			log.Error("Get commits", err)
		}
	}

	s.StartAsync()
	s.StartBlocking()
}
