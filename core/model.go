package model

import (
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"time"
)

type Messages struct {
	From    string `json:"from"`
	GoTo    string `json:"go_to"`
	MsgType int    `json:"msg_type"`
	Content string `json:"content"`
}
type User struct {
	ChineseName string `json:"chinese_name"`

	Token    string `json:"token"`
	LastTime *time.Time
	Conn     *websocket.Conn
}

func NewUser(conn *websocket.Conn, name string) User {
	now := time.Now()
	user := User{
		Token:       uuid.NewV4().String(),
		LastTime:    &now,
		Conn:        conn,
		ChineseName: name,
	}
	return user
}
