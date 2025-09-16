package log

import (
	"github.com/sirupsen/logrus"
	"time"
)

// Config 日志配置
// example:
// {
//	"name": "example",
// 	"level": 5,
//	"json": false,
// 	"outputs": {
// 		"file": {
// 			"path": "./logs"
//			"rotate": true,
//			"json": true,
// 		},
// 		"logstash": {
// 			"addr": "localhost:5000",
// 		}
// 	}
// }

type IConfig interface {
	ParseOptions() []Option
}
type Config struct {
	Name    string `json:"name"`
	Level   int    `json:"level"`
	UseJSON bool   `json:"json"`
	RootDir string `json:"rootdir"`

	// 必定会输出到console
	OutputFile      bool   `json:"outfile"`
	FilePath        string `json:"filepath"`
	OutputLogstash  bool   `json:"outstash"`
	LogstashAddr    string `json:"stashaddr"`
	LogStashNetWork string `json:"stashnetwork"`
}

func (c *Config) ParseOptions() []Option {
	var ops = []Option{
		WithName(c.Name),
		WithLevel(logrus.Level(c.Level)),
		WithUseJSON(c.UseJSON),
		WithRootDir(c.RootDir),
		WithOutputFile(c.OutputFile),
		WithFilePath(c.FilePath),
		WithOutputLogstash(c.OutputLogstash),
		WithLogstashAddr(c.LogstashAddr),
		WithLogstashNetWork(c.LogStashNetWork),
	}
	return ops
}

var globalConfig = NewConfig()

type defaultConfig struct {
	opts *options
}

func NewConfig(opts ...Option) *defaultConfig {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	if o.OutputFile && o.FilePath == "" {
		panic("log output file path is required")
	}
	if o.OutputLogstash {
		if o.LogstashAddr == "" {
			panic("log stash addr is required")
		} else if o.LogStashNetWork != "ws" && o.LogStashNetWork != "tcp" {
			panic("log stash network is required, only support ws or tcp")
		}
	}

	c := &defaultConfig{}
	c.opts = o
	return c
}

// InitLogrus 根据配置初始化 logrus, 添加配置的 Hooks
func InitLogrus(opts ...Option) (close func(), err error) {
	if len(opts) > 0 {
		globalConfig = NewConfig(opts...)
	}
	logrus.SetLevel(logrus.Level(globalConfig.opts.Level))
	if globalConfig.opts.UseJSON {
		logrus.SetFormatter(new(logrus.JSONFormatter))
	} else {
		text := new(logrus.TextFormatter)
		text.FullTimestamp = true
		logrus.SetFormatter(text)
	}
	logrus.SetReportCaller(true)
	return addLogHooks(logrus.StandardLogger(), globalConfig)
}

// Logger ...
func Logger() *logrus.Logger {
	return logrus.StandardLogger()
}

func addLogHooks(logger *logrus.Logger, cfg *defaultConfig) (close func(), err error) {
	var fcloses []func()
	defer func() {
		if err != nil {
			for _, f := range fcloses {
				f()
			}
		}
	}()
	if cfg.opts.OutputFile {
		close, err = addFileHook(logger, cfg)
		if err != nil {
			return nil, err
		}
		fcloses = append(fcloses, close)
	}
	if cfg.opts.OutputLogstash {
		close, err = addLogstashHook(logger, cfg)
		if err != nil {
			return nil, err
		}
		fcloses = append(fcloses, close)
	}
	return func() {
		for _, f := range fcloses {
			f()
		}
	}, nil
}

func addFileHook(logger *logrus.Logger, cfg *defaultConfig) (close func(), err error) {
	var path string

	if len(cfg.opts.FilePath) > 0 {
		path = cfg.opts.FilePath
	} else {
		path = cfg.opts.Name
	}

	hook, close, err := NewFileLogHook(path, cfg.opts.Name, cfg.opts.UseJSON, true)
	if err != nil {
		return nil, err
	}

	logger.AddHook(hook)

	return close, nil
}

func addLogstashHook(logger *logrus.Logger, cfg *defaultConfig) (closef func(), err error) {
	addr := cfg.opts.LogstashAddr

	log := logger.WithField("addr", addr)

	log.Info("Connecting logstash...")
	hook, inclosef, err := NewLogstashHook(addr, cfg.opts.LogStashNetWork, cfg.opts.Name, cfg.opts.Name, cfg.opts.RootDir)
	if err != nil {
		log.WithError(err).Error("Connect logstash failed")
		return nil, err
	}
	log.Info("Connect logstash succ")

	logger.AddHook(hook)

	ch := make(chan struct{})
	if cfg.opts.LogStashNetWork == "ws" {
		go func() {
			ticker := time.NewTicker(time.Second * 10)
			for {
				select {
				case <-ch:
					return
				case <-ticker.C:
					log.Trace("heartbeat")
				}
			}
		}()
	}

	cf := func() {
		close(ch)
		inclosef()
	}

	return cf, nil
}
