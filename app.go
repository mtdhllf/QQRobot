package main

import (
	"encoding/json"
	"github.com/Tnze/CoolQ-Golang-SDK/cqp"
	"github.com/Tnze/CoolQ-Golang-SDK/cqp/util"
	"github.com/json-iterator/go"
	"github.com/robfig/cron"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

//go:generate cqcfg -c .
// cqp: 名称: QQRobot
// cqp: 版本: 1.0.0:1
// cqp: 作者: Gopher
// cqp: 简介: 一个超棒的Go语言插件Demo，它会回复你的私聊消息~
func main() { /*此处应当留空*/ }

const RobotUrl = "http://api.qingyunke.com/api.php?key=free&appid=0&msg="

var err error = nil

var c = cron.New()

func init() {
	cqp.AppID = "me.qqbobot.demo" // TODO: 修改为这个插件的ID
	cqp.PrivateMsg = onPrivateMsg
	cqp.GroupMsg = onGroupMsg
	//定时任务
	initJob()
}

func onPrivateMsg(subType, msgID int32, fromQQ int64, msg string, font int32) int32 {
	cqp.SendPrivateMsg(fromQQ, msg) //复读机
	return 0
}

//群聊入口
func onGroupMsg(subType, msgID int32, fromGroup, fromQQ int64, fromAnonymous, msg string, font int32) int32 {
	//是否@消息
	if hasAtSelf(msg) {
		//智能聊天
		robotAnswer(fromGroup, fromQQ, msg)
		//1表示已处理消息,不再传递给其他插件处理
		return 1
	}
	//处理其他消息
	code := onKeyGroupMsg(subType, msgID, fromGroup, fromQQ, fromAnonymous, msg, font)
	return code
}

//是否@自己
func hasAtSelf(msg string) bool {
	reg := regexp.MustCompile(`\[CQ:at,qq=(\d+)\]`)
	match := reg.FindStringSubmatch(msg)
	for _, v := range match {
		if strconv.FormatInt(cqp.GetLoginQQ(), 10) == v {
			return true
		}
	}
	return false
}

//机器人智能回复
func robotAnswer(fromGroup, fromQQ int64, msg string) {
	//get请求
	//http.Get的参数必须是带http://协议头的完整url,不然请求结果为空
	cqp.AddLog(cqp.Debug, "robotAnswer-msg", msg)
	resp, _ := http.Get(RobotUrl + msg)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	cqp.AddLog(cqp.Debug, "robotAnswer-body", string(body))
	var robotMsg RobotMsg
	var jsonIterator = jsoniter.ConfigCompatibleWithStandardLibrary
	if err := jsonIterator.Unmarshal(body, &robotMsg); err == nil {
		cqp.SendGroupMsg(fromGroup, util.CQCode("at", "qq", fromQQ)+util.Escape(robotMsg.Content))
	} else {
		cqp.AddLog(cqp.Debug, "robotAnswer-answer", err.Error())
	}
}

//消息处理
func onKeyGroupMsg(subType, msgID int32, fromGroup, fromQQ int64, fromAnonymous, msg string, font int32) int32 {
	code := int32(0)
	//一个搜题功能,不知接口是否还能正常工作
	if strings.HasPrefix(msg, "搜题 ") && len(strings.Split(msg, " ")) > 1 {
		split := strings.Split(msg, " ")
		cqp.AddLog(cqp.Debug, "搜题", split[1])
		resp, _ := http.Post("https://ninja.yua.im/ninja/qa",
			"application/x-www-form-urlencoded",
			strings.NewReader("search="+split[1]))
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		response := string(body)
		msg := "搜题出错,请稍后再试~"

		var exam Exam
		if err := json.Unmarshal([]byte(response), &exam); err == nil {
			if exam.IsSuc && exam.Data.Total > 0 {
				all := make([]string, exam.Data.Total+1)
				all[0] = "小改改为你找到以下结果:"
				for index, v := range exam.Data.Rows {
					title := strconv.Itoa(index+1) + ". " + v.Title
					s1 := make([]string, len(v.Answers)+1)
					s1[0] = title
					for k, v := range v.Answers {
						if v.IsCorrect {
							//正确
							s1[k+1] = v.Content + " √"
						} else {
							s1[k+1] = v.Content
						}
					}
					//一条题目及回答
					all[index+1] = strings.Join(s1, "\n")
				}
				msg = strings.TrimSpace(strings.Join(all, "\n"))
			} else {
				msg = "没有找到你要的题目哦~"
			}
		}
		cqp.SendGroupMsg(fromGroup, msg)
	}
	//帮助
	if strings.Contains(msg, "帮助") {
		code = 1
		cqp.SendGroupMsg(fromGroup, "帮助菜单:\n"+"巴拉巴拉~")
	}
	return code
}

//定时任务
func initJob() {
	//早晨播报
	err = c.AddFunc("5 0 7 * * ?", func() {
		cqp.SendGroupMsg(816440954, "早上好,今天也是充满希望的一天(●'◡'●)ﾉ")
	})
	//晚上播报
	err = c.AddFunc("5 0 23 * * ?", func() {
		cqp.SendGroupMsg(816440954, "【碎觉碎觉】")
	})

	if err != nil {
		cqp.AddLog(cqp.Debug, "job", err.Error())
		return
	}
	c.Start()
}

//机器人智能回复消息
type RobotMsg struct {
	Result  int    `json:"result"`
	Content string `json:"content"`
}

//搜题
type Exam struct {
	Data struct {
		Rows []struct {
			_id     string `json:"_id"`
			Agree   int    `json:"agree"`
			Answers []struct {
				Content   string `json:"content"`
				IsCorrect bool   `json:"is_correct"`
			} `json:"answers"`
			CreateDt  string `json:"create_dt"`
			CreatedBy string `json:"created_by"`
			Disagree  int    `json:"disagree"`
			Tip       string `json:"tip"`
			Title     string `json:"title"`
			UpdateDt  string `json:"update_dt"`
			UpdatedBy string `json:"updated_by"`
		} `json:"rows"`
		Total int `json:"total"`
	} `json:"data"`
	IsSuc bool `json:"is_suc"`
}
