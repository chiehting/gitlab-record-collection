package services

import (
	"sort"

	"github.com/chiehting/gitlab-record-collection/pkg/config"
	"github.com/chiehting/gitlab-record-collection/pkg/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	cloudwatchlogs "github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

type awsClient struct {
	session       *session.Session
	cloudwatchLog *cloudwatchlogs.CloudWatchLogs
}

// AWS Client
var AWS awsClient
var awsConfig = config.GetAWS()

// CreateSession 初始化 AWS session，使用你的 AWS credentials 和 region
func (awsClient *awsClient) CreateSession() {
	session, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-1"),
	})
	if err != nil {
		log.Panic("無法建立 AWS session：", err)
	}

	awsClient.session = session
}

// 建立 CloudWatch log 的 service client
func (awsClient *awsClient) ConnectionCloudwatchLog() {
	session := awsClient.session
	awsClient.cloudwatchLog = cloudwatchlogs.New(session)
}

// 建立 Log stream
func (awsClient *awsClient) CreateLogStream(logGroupName string, logStreamName string) {
	paramsStream := &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
	}

	// 執行 CreateLogStream 請求
	_, err := awsClient.cloudwatchLog.CreateLogStream(paramsStream)
	if err != nil {
		// 檢查錯誤類型是否是 "ResourceAlreadyExistsException"
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "ResourceAlreadyExistsException" {
				log.Warn("日誌流已經存在，無需再次建立。")
			} else {
				log.Warn("無法建立日誌流：", err)
			}
		} else {
			log.Warn("無法建立日誌流：", err)
		}
	} else {
		log.Info("日誌流已成功建立。")
	}
}

// 建立 PutLogEvents 請求
func (awsClient *awsClient) PutLogEvents(logGroupName string, logStreamName string, logEvents []*cloudwatchlogs.InputLogEvent) {
	sort.Slice(logEvents, func(i, j int) bool {
		return *logEvents[i].Timestamp < *logEvents[j].Timestamp
	})

	params := &cloudwatchlogs.PutLogEventsInput{
		LogEvents:     logEvents,
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
	}

	// 執行 PutLogEvents 請求
	_, err := awsClient.cloudwatchLog.PutLogEvents(params)
	if err != nil {
		log.Error("無法傳送日誌至 CloudWatch：", err)
	}
}
