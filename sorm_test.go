package sorm

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gongshen/sorm"
	"testing"
)

const (
	dsn   = "root:123456@tcp(127.0.0.1:3306)/orm_db?charset=utf8&loc=Local"
	table = "user"
)

type User struct {
	ID        int64  `json:"id"`
	Age       int64  `json:"age"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func TestQuery_Insert(t *testing.T) {
	ormDB, _ := sorm.Connect(dsn)
	users := sorm.Table(ormDB, table)

	user1 := &User{
		Age:       11,
		FirstName: "Tom",
		LastName:  "One",
	}
	user2 := User{
		Age: 12,
	}
	user3 := map[string]interface{}{
		"age":       13,
		"FirstName": "Tom",
		"LastName":  "Three",
	}
	_, err := users().Insert([]interface{}{user1, user2})
	_, err = users().Insert(user3)
	if err != nil {
		fmt.Printf("插入失败！Err：%v", err)
	} else {
		fmt.Printf("插入成功！")
	}
}

func TestQuery_Select(t *testing.T) {
	ormDB, _ := sorm.Connect(dsn)
	users := sorm.Table(ormDB, table)
	var u User
	err := users().Where(&User{FirstName:"Tom",Age:20}).Only("age","first_name","last_name").Select(&u)
	if err != nil {
		fmt.Printf("查询失败！Error：%v", err)
	} else {
		fmt.Println("查询成功！")
	}
}

func TestQuery_Update(t *testing.T) {
	ormDB, _ := sorm.Connect(dsn)
	users := sorm.Table(ormDB, table)
	u1 := "age = 100"
	u2 := map[string]interface{}{
		"age":        20,
		"first_name": "Tom",
		"last_name":  "gong",
	}
	u3 := &User{
		Age:       56,
		FirstName: "z",
		LastName:  "zy",
	}
	_, _ = users().Where("age > 10").Update(u1)
	_, _ = users().Where("age > 10").Update(u2)
	_, _ = users().Where("age =30").Update(u3)
}
