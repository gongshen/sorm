package sorm

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gongshen/sorm"
	"testing"
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

func TestQuery_Select(t *testing.T) {
	ormDB, _ := orm.Connect("root:123456@tcp(127.0.0.1:3306)/orm_db?charset=utf8&loc=Local")
	users := orm.Table(ormDB, "user")
	var user User
	err := users().Only("created_at").Select(&user)
	if err != nil {
		fmt.Println("查询失败！")
	} else {
		fmt.Println("查询成功！")
	}
}

func TestQuery_Update(t *testing.T) {
	ormDB,_:=orm.Connect("root:123456@tcp(127.0.0.1:3306)/orm_db?charset=utf8")
	users:=orm.Table(ormDB,"user")
	u1:="age=100"
	u2:=map[string]interface{}{
		"age":"10",
		"first_name":"Jim",
		"last_name":"Three",
	}
	u3:=&User{
		Age:30,
		FirstName:"Janny",
		LastName:"Ban",
	}
	users().Where("id = 2").
}

