package config

import (
	"strings"

	"github.com/spf13/viper"
)

const (
	_configFile      = "config/config.yml"
	_configEnvPrefix = "env"
)

// Application is define application
type Application struct {
	Name    string
	Version string
}

// GitLab is
type GitLab struct {
	Scheme string
	Domain string
	Token  string
}

// AWS is setting http server
type AWS struct {
	LogGroupName  string
	LogStreamName string
}

// Log is setting log
type Log struct {
	Level       string
	FilePath    string
	OmitTimeKey bool
}

// LevelDB setting
type LevelDB struct {
	DBPath string
}

type config struct {
	Application Application
	GitLabs     []GitLab
	AWS         AWS
	Log         Log
	LevelDB     LevelDB
}

var cfg *config

// Setup initialize the configuration instance
func init() {
	setConfigDefault()
	err := getConfig()
	if err != nil {
		panic(err)
	}
}

func setConfigDefault() {
	var null interface{}
	// application 預設值
	viper.SetDefault("application.name", "")

	// log 預設值
	viper.SetDefault("log.level", "debug")
	viper.SetDefault("log.filePath", null)
	viper.SetDefault("log.omittimekey", false)

	// levelDB 預設值
	viper.SetDefault("leveldb.dbPath", "./.leveldb")
}

func getConfig() (err error) {
	viper.AutomaticEnv()
	viper.SetEnvPrefix(_configEnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetConfigFile(_configFile)
	if err := viper.ReadInConfig(); err == nil {
		viper.Unmarshal(&cfg)
	}

	cfg.prepareConfig()

	return
}

// prepareConfig 調整設定檔
func (cfg *config) prepareConfig() {
	cfg.Log.Level = strings.ToLower(cfg.Log.Level)

	if len(cfg.Application.Name) == 0 {
		panic("Please set the application name to config file or environment variable!")
	}
}

// GetApplication 取得服務資訊
func GetApplication() *Application {
	return &cfg.Application
}

// GetGitLabs is
func GetGitLabs() *[]GitLab {
	return &cfg.GitLabs
}

// GetLevelDB get LevelDB setting
func GetLevelDB() *LevelDB {
	return &cfg.LevelDB
}

// GetLog 取得日誌資訊
func GetLog() *Log {
	return &cfg.Log
}

// GetAWS 取得AWS資訊
func GetAWS() *AWS {
	return &cfg.AWS
}
