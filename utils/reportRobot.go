package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

var robot IReportRobot

type IReportRobot interface {
	ReportToRobotChat(chatmsg string) error
}

type RocketRobot struct {
	hook string
}

func (ro *RocketRobot) ReportToRobotChat(content string) error {
	if ro.hook == "" {
		return fmt.Errorf("hook is empty")
	}
	type ReqData struct {
		Text string `json:"text"`
	}

	req := ReqData{Text: content}

	reqbody, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(ro.hook, "application/json", strings.NewReader(string(reqbody)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println("返回：", string(result))
	return nil
}

func GetReportRobotIns() IReportRobot {
	if robot == nil {
		robot = new(RocketRobot)
	}
	return robot
}

func RegisterRobot(typ string, hookurl string) {
	if robot == nil {
		robot = &RocketRobot{hook: hookurl}
	}
}
