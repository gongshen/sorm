//数据库的更新操作
package sorm

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"reflect"
	"strings"
)

//update user set first_name = "z", last_name = "zy" where first_name = "Tom" and last_name = "Curise"
func (q *Query)Update(in interface{})(int64,error){
	if len(q.errs)!=0{
		return 0,errors.New(strings.Join(q.errs," "))
	}
	v:=reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr{
		v=v.Elem()
	}
	var where,toBeUpdate string
	var keys,values []string
	switch v.Kind() {
	case reflect.String:
		toBeUpdate = in.(string)
	case reflect.Struct:
		keys,values=sKV(v)
	case reflect.Map:
		keys,values=mKV(v)
	default:
		return 0,errors.New("Method Update:Type error!")
	}
	if toBeUpdate==""{
		if len(keys) != len(values){
			return 0,errors.New("Method Update:Key not match Values!")
		}
		var tmp []string
		for i,k:=range keys{
			tmp=append(tmp,fmt.Sprintf("%s = %s",k,values[i]))
		}
		toBeUpdate=strings.Join(tmp,",")
	}
	if len(q.wheres)>0{
		where =fmt.Sprintf(`where %s`,strings.Join(q.wheres," and "))
	}
	query:=fmt.Sprintf("update %s set %s %s",q.table,where,toBeUpdate)
	log.Printf(`update sql:%s`,query)
	st,err:=q.db.Prepare(query)
	if err!=nil{
		return 0,err
	}
	result,err:=st.Exec()
	if err!=nil{
		return 0,err
	}
	return result.RowsAffected()
}


