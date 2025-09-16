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

// InitLogrus 根据配置初始化 logrus, 添加配置的 Hooks
func InitLogrus(cfg *Config) (close func(), err error) {
	logrus.SetLevel(logrus.Level(cfg.Level))
	if cfg.UseJSON {
		logrus.SetFormatter(new(logrus.JSONFormatter))
	} else {
		text := new(logrus.TextFormatter)
		text.FullTimestamp = true
		logrus.SetFormatter(text)
	}
	logrus.SetReportCaller(true)
	return addLogHooks(logrus.StandardLogger(), cfg)
}

// Logger ...
func Logger() *logrus.Logger {
	return logrus.StandardLogger()
}

func addLogHooks(logger *logrus.Logger, cfg *Config) (close func(), err error) {
	var fcloses []func()
	defer func() {
		if err != nil {
			for _, f := range fcloses {
				f()
			}
		}
	}()
	if cfg.OutputFile {
		close, err = addFileHook(logger, cfg)
		if err != nil {
			return nil, err
		}
		fcloses = append(fcloses, close)
	}
	if cfg.OutputLogstash {
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

func addFileHook(logger *logrus.Logger, cfg *Config) (close func(), err error) {
	var path string

	if len(cfg.FilePath) > 0 {
		path = cfg.FilePath
	} else {
		path = cfg.Name
	}

	hook, close, err := NewFileLogHook(path, cfg.Name, cfg.UseJSON, true)
	if err != nil {
		return nil, err
	}

	logger.AddHook(hook)

	return close, nil
}

func addLogstashHook(logger *logrus.Logger, cfg *Config) (closef func(), err error) {
	addr := cfg.LogstashAddr

	log := logger.WithField("addr", addr)

	log.Info("Connecting logstash...")
	hook, inclosef, err := NewLogstashHook(addr, cfg.LogStashNetWork, cfg.Name, cfg.Name, cfg.RootDir)
	if err != nil {
		log.WithError(err).Error("Connect logstash failed")
		return nil, err
	}
	log.Info("Connect logstash succ")

	logger.AddHook(hook)

	ch := make(chan struct{})
	if cfg.LogStashNetWork == "ws" {
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
