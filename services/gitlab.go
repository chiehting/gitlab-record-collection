package services

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/chiehting/gitlab-record-collection/pkg/config"
	"github.com/chiehting/gitlab-record-collection/pkg/log"
)

type gitlab struct{}

// Project is value from GitLab Project API
type Project struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Path              string `json:"path"`
	PathWithNamespace string `json:"path_with_namespace"`
	LastActivityAt    string `json:"last_activity_at"`
}

// Commit is value from GitLab Commits API
type Commit struct {
	ShortID       string `json:"short_id"`
	Title         string `json:"title"`
	AuthorName    string `json:"author_name"`
	AuthorEmail   string `json:"author_email"`
	CommittedDate string `json:"committed_date"`
	Message       string `json:"message"`
}

// CommitWithProject 是新的結構，包含 Commit 資料以及專案欄位
type CommitWithProject struct {
	Commit      Commit `json:"commit"`
	ProjectID   int    `json:"project_id"`
	ProjectName string `json:"project_name"`
}

// Gitlab is to provide settings configuration
var Gitlab *gitlab

// GetProject is initialization when the service started
func (gitlab *gitlab) GetProjects(target config.GitLab) []Project {

	GitLabURL := target.Scheme + target.Domain + "/api/v4/projects?archived=false&order_by=last_activity_at&sort=desc"
	GitLabToken := target.Token

	// 發送 GET 請求
	req, err := http.NewRequest("GET", GitLabURL, nil)
	if err != nil {
		log.Error("發送 GET 請求失敗：", err)
		return nil
	}
	req.Header.Add("PRIVATE-TOKEN", GitLabToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("接收回應失敗：", err)
		return nil
	}
	defer resp.Body.Close()

	// 讀取回應內容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("讀取回應內容失敗：", err)
		return nil
	}

	// 解析 JSON 資料
	var projects []Project
	err = json.Unmarshal(body, &projects)
	if err != nil {
		log.Error("解析 JSON 失敗：", err)
		return nil
	}

	return projects
}

// GetProjectList is initialization when the service started
func (gitlab *gitlab) GetCommits(target config.GitLab) []CommitWithProject {

	projects := gitlab.GetProjects(target)
	gitLabToken := target.Token
	var commitsWithProject []CommitWithProject
	var commits []Commit

	currentTime := time.Now()
	oneHourAgo := currentTime.Add(-time.Hour)
	timeLayout := "2006-01-02T3:04:05Z"
	formattedTime := oneHourAgo.Format(timeLayout)

	for _, project := range projects {
		GitLabURL := target.Scheme + target.Domain + "/api/v4/projects/" + strconv.Itoa(project.ID) + "/repository/commits?since=" + formattedTime
		log.Debug(GitLabURL)
		req, err := http.NewRequest("GET", GitLabURL, nil)
		if err != nil {
			log.Error("發送 GET 請求失敗：", err)
			return nil
		}
		req.Header.Add("PRIVATE-TOKEN", gitLabToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Error("接收回應失敗：", err)
			return nil
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error("讀取回應內容失敗：", err)
			return nil
		}

		err = json.Unmarshal(body, &commits)
		if err != nil {
			log.Error("解析 JSON 失敗：", err)
			return nil
		}

		for _, commit := range commits {
			commitWithProject := CommitWithProject{
				Commit:      commit,
				ProjectID:   project.ID,
				ProjectName: project.Name,
			}
			commitsWithProject = append(commitsWithProject, commitWithProject)
		}

	}

	return commitsWithProject
}
