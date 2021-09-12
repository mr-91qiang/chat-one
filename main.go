package main

import (
	model "char/core"
	"char/errs"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"os"
)

type Rsp struct {
	Code int    `json:"Code"`
	Msg  string `json:"msg"`
	Data string
}

const (
	System = "System"
	NAMING = 9
)

func init() {
	logFile, err := os.OpenFile("./ws.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile) // 将文件设置为log输出的文件
	log.SetPrefix("TRACE: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)
}

type Friend struct {
	Token string
	Name  string
}

// ConnMCap 包含全部用户信息
var ConnMCap = make(map[string]*model.User)

var closeChan = make(chan string, 2)

func main() {
	closeChan = make(chan string, 100)
	go removeUser(closeChan)
	http.HandleFunc("/ws", Upgrade)
	http.HandleFunc("/getFriend", GetFriend)
	log.Println("准备启动")
	err := http.ListenAndServe(":5900", nil)
	if err != nil {
		log.Println(err)
		return
	}
}
func removeUser(closeChan chan string) {
	for {
		token := <-closeChan
		log.Println(token)
		if user, ok := ConnMCap[token]; ok {
			delete(ConnMCap, token)
			log.Printf("删除成功token:%s,name:%s\n", token, user.ChineseName)
			continue
		}
		log.Printf("删除失败token:%s 不存在\n", token)
	}
}
func GetFriend(writer http.ResponseWriter, _ *http.Request) {
	rsp := Rsp{}
	users := getUsers()
	marshal, err := json.Marshal(users)
	if err != nil {
		return
	}
	_, err = io.WriteString(writer, string(marshal))
	if err != nil {
		return
	}
	rsp.Code = 20000
	rsp.Msg = "success"
	rsp.Data = string(marshal)

}

func Upgrade(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	OnLine(conn)
}

//获取全部用户
func getUsers() []Friend {
	var friends = make([]Friend, 0, len(ConnMCap))
	for s, user := range ConnMCap {
		friends = append(friends, Friend{
			Token: s,
			Name:  user.ChineseName,
		})
	}
	return friends
}
func OnLine(conn *websocket.Conn) {
	user := model.NewUser(conn, "")
	log.Println("接入用户:", user)
	ConnMCap[user.Token] = &user
	go From(&user)
}

func From(user *model.User) {
	for {
		//var msg model.Mess
		messageType, str, err := user.Conn.ReadMessage()
		if err != nil {
			//如果连接关闭,删除map中的值
			if _, ok := err.(*websocket.CloseError); ok {
				closeChan <- user.Token
				return
			}
		}
		switch messageType {
		case websocket.CloseMessage:
			closeChan <- user.Token
			log.Println("连接关闭")
		case websocket.TextMessage:
			err = FilterNoName(user.ChineseName, str)
			if err != nil {
				go To(System, user.Token, errs.Msg(err))
				continue
			}
			log.Println(string(str))
		case -1:
			log.Println("连接关闭")
			closeChan <- user.Token
			return
		default:
			log.Println("接收到无意义数据:", string(str))
		}
	}
}
func To(from, to, msg string) {
	var fromChineseName string
	if System == from {
		fromChineseName = System
	} else if fromP, is := ConnMCap[from]; is {
		fromChineseName = fromP.ChineseName
	} else {
		log.Println("用户不存在")
	}
	toP, is2 := ConnMCap[to]
	//這裡先不處理用戶下線的情況
	if !is2 {
		log.Println("用户不存在")
		return
	}
	messages := model.Messages{
		From:    fromChineseName,
		GoTo:    toP.ChineseName,
		Content: msg,
	}
	err := toP.Conn.WriteJSON(messages)
	log.Println(err)
}
func FilterNoName(token string, bytes []byte) error {
	var fromP = ConnMCap[token]
	if fromP == nil {
		return errs.CustomError{
			Code: 400,
			Msg:  "用户未登陆",
		}
	}
	var msg model.Messages
	err := json.Unmarshal(bytes, &msg)
	if err != nil {
		return errs.CustomError{
			Code: 400,
			Msg:  fmt.Sprint("消息格式有误:", err.Error()),
		}
	}
	if msg.MsgType == NAMING {
		fromP.ChineseName = msg.Content
	}
	if fromP.ChineseName == "" {
		return errs.CustomError{
			Code: 400,
			Msg:  "用户未登陆,请先填写你的昵称！！",
		}
	}
	return nil
}

func ToUser(from string, to model.User, msg string) {
	var fromChineseName string
	if System == from {
		fromChineseName = System
	} else if fromP, is := ConnMCap[from]; is {
		fromChineseName = fromP.ChineseName
	} else {
		log.Println("用户不存在")
	}
	messages := model.Messages{
		From:    fromChineseName,
		GoTo:    to.ChineseName,
		Content: msg,
	}
	err := to.Conn.WriteJSON(messages)
	log.Println(err)
}
