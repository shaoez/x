package log

import "github.com/sirupsen/logrus"

const (
	defaultName            = "log"
	defaultLevel           = logrus.TraceLevel
	defaultUseJson         = true
	defaultRootDir         = "."
	defaultOutputFile      = true
	defaultFilePath        = "./log"
	defaultOutPutLogstash  = false
	defaultLogstashAddr    = "127.0.0.1:52001"
	defaultLogStashNetWork = "ws"
)

type Option func(o *options)

type options struct {
	Name    string       `json:"name"`
	Level   logrus.Level `json:"level"`
	UseJSON bool         `json:"json"`
	RootDir string       `json:"rootdir"`

	// 必定会输出到console
	OutputFile      bool   `json:"outfile"`
	FilePath        string `json:"filepath"`
	OutputLogstash  bool   `json:"outstash"`
	LogstashAddr    string `json:"stashaddr"`
	LogStashNetWork string `json:"stashnetwork"`
}

func defaultOptions() *options {
	this := &options{
		Name:            defaultName,
		Level:           defaultLevel,
		UseJSON:         defaultUseJson,
		RootDir:         defaultRootDir,
		OutputFile:      defaultOutputFile,
		FilePath:        defaultFilePath,
		OutputLogstash:  defaultOutputFile,
		LogstashAddr:    defaultLogstashAddr,
		LogStashNetWork: defaultLogStashNetWork,
	}

	return this
}

func WithName(name string) Option {
	return func(o *options) { o.Name = name }
}

func WithLevel(level logrus.Level) Option {
	return func(o *options) { o.Level = level }
}

func WithUseJSON(useJSON bool) Option {
	return func(o *options) { o.UseJSON = useJSON }
}

func WithRootDir(rootDir string) Option {
	return func(o *options) { o.RootDir = rootDir }
}

func WithOutputFile(outputFile bool) Option {
	return func(o *options) { o.OutputFile = outputFile }
}

func WithFilePath(path string) Option {
	return func(o *options) {
		o.FilePath = path
	}
}

func WithOutputLogstash(outputLogstash bool) Option {
	return func(o *options) { o.OutputLogstash = outputLogstash }
}

func WithLogstashAddr(logstashAddr string) Option {
	return func(o *options) { o.LogstashAddr = logstashAddr }
}

func WithLogstashNetWork(logstashNetWork string) Option {
	return func(o *options) { o.LogStashNetWork = logstashNetWork }
}
