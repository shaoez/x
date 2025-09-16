package logstash

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"net"
	"net/url"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	logrustash "github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/sirupsen/logrus"
)

//日志配置
type GameLog struct {
	FileName    string `json:"fileName"`
	Level       int32  `json:"level"`
	ChannelName string `json:"channelName"`
	LogType     int32  `json:"logType"`
	Msg         string `json:"msg"`
}

var wsLeveMap map[logrus.Level]int32 = map[logrus.Level]int32{
	logrus.PanicLevel: 3,
	logrus.FatalLevel: 3,
	logrus.ErrorLevel: 3,
	logrus.WarnLevel:  4,
	logrus.InfoLevel:  6,
	logrus.DebugLevel: 7,
	logrus.TraceLevel: 8,
}

// Hook logstash hook for logrus
type Hook struct {
	conn     net.Conn
	impl     logrus.Hook
	ch       chan int
	closed   int32
	network  string
	fileName string
	rootDir  string

	rw sync.RWMutex
}

// New 创建一个 logstash hook
func New(addr string, fields logrus.Fields) (*Hook, error) {
	var path string = strings.Split(addr, ":")[1]
	network := fields["network"].(string)
	name := fields["name"].(string)
	rootDir := fields["rootdir"].(string)
	conn, err := dial(network, addr, path)
	if err != nil {
		return nil, err
	}

	h := new(Hook)
	h.conn = conn
	h.network = network
	h.fileName = name
	h.rootDir = rootDir
	h.impl, err = logrustash.NewHookWithConn(conn, addr)
	if err != nil {
		return nil, err
	}

	h.ch = make(chan int)

	go func() {
		for range h.ch {
			var conn net.Conn
			for {
				var err error
				conn, err = dial(network, addr, path)
				if err == nil {
					break
				}
				if atomic.LoadInt32(&h.closed) != 0 {
					return
				}
				time.Sleep(0)
			}
			impl, _ := logrustash.NewHookWithConn(conn, addr)

			if atomic.LoadInt32(&h.closed) != 0 {
				return
			}

			h.rw.Lock()
			h.conn = conn
			h.impl = impl
			h.rw.Unlock()
		}
	}()

	return h, nil
}

func dial(network string, addr string, path string) (net.Conn, error) {
	if network == "ws" {
		u := url.URL{Scheme: "ws", Host: addr, Path: path}
		return websocket.Dial(u.String(), "", "*")

	} else if network == "tcp" {
		return net.Dial("tcp", addr)
	}
	return nil, fmt.Errorf("network %s not support", network)
}

// Levels logrus.Hook interface
func (h *Hook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	}
}

// Fire logrus.Hook interface
func (h *Hook) Fire(entry *logrus.Entry) error {
	h.rw.RLock()
	defer h.rw.RUnlock()
	var err error
	if h.network == "ws" {
		var bts []byte
		glog := new(GameLog)
		glog.Msg = h.format(entry)
		glog.FileName = h.fileName
		glog.Level = wsLeveMap[entry.Level]

		bts, err = json.Marshal(glog)
		if err != nil {
			return err
		}
		_, err = h.conn.Write(bts)
	} else {
		err = h.impl.Fire(entry)
	}

	if err != nil {
		h.conn.Close()
		select {
		case h.ch <- 0:
		default:
		}
	}
	return err
}

// Close 关闭连接
func (h *Hook) Close() error {
	if atomic.CompareAndSwapInt32(&h.closed, 0, 1) {
		h.rw.RLock()
		defer h.rw.RUnlock()
		close(h.ch)
		return h.conn.Close()
	}
	return nil
}

func (h *Hook) format(entry *logrus.Entry) string {

	var msg string
	var err error

	fields := make(map[string]string)
	fields["@time"] = entry.Time.Format("2006-01-02 15:04:05")
	fields["@caller"] = ""
	if entry.HasCaller() {
		if h.rootDir != "" {
			var dir string
			dirs := strings.Split(entry.Caller.File, h.rootDir)
			if len(dirs) > 1 {
				dir = dirs[1]
			}
			fields["@caller"] = fmt.Sprintf("%s:%s:%d", h.fileName, dir, entry.Caller.Line)
		} else {
			dir, filename := path.Split(entry.Caller.File)
			dirall := strings.Split(dir, "/")
			dir = dirall[len(dirall)-2]
			fields["@caller"] = fmt.Sprintf("%s:%s/%s:%d", h.fileName, dir, filename, entry.Caller.Line)
		}
	}
	msg = fmt.Sprintf("[%s]|[%s] | ", fields["@time"], fields["@caller"])
	var bts []byte
	bts, err = json.Marshal(entry.Data)
	if err == nil && len(bts) > 0 {
		msg = fmt.Sprintf("%s[%s] | ", msg, string(bts))
	}

	if len(entry.Message) > 0 {
		msg = fmt.Sprintf("%s[%s]", msg, entry.Message)
	}

	return msg + "\n"
}
