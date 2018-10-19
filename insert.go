//数据库的插入操作
package sorm

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
)

//insert into user (age,first_name,last_name) values (20,'Tom','One')
func (q *Query) Insert(in interface{}) (int64, error) {
	var keys, values []string
	v := reflect.ValueOf(in)
	//剥离指针
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	//判断in的类型
	switch v.Kind() {
	case reflect.Struct:
		keys, values = sKV(v)
	case reflect.Map:
		keys, values = mKV(v)
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			sv := v.Index(i)
			//剥离指针
			for sv.Kind() == reflect.Interface || sv.Kind() == reflect.Ptr {
				sv = sv.Elem()
			}
			//&{"Tom","One"}   {"Tom","Second"}
			if sv.Kind() != reflect.Struct {
				return 0, errors.New("method Insert error: in slices is not struct!")
			}
			if len(keys) == 0 {
				keys, values = sKV(sv)
				continue
			}
			//key保存一次就行了
			_, val := sKV(sv)
			values = append(values, val...)
		}
	default:
		return 0, errors.New("method Insert error: type error!")
	}
	keysLen := len(keys)
	valuesLen := len(values)
	if keysLen == 0 || valuesLen == 0 {
		return 0, errors.New("method Insert error：no data!")
	}
	var insertValue string
	//插入多条记录时，使用 "," 拼接values
	if keysLen < valuesLen {
		var tmpValues []string
		for keysLen <= valuesLen {
			if keysLen%(len(keys)) == 0 {
				tmpValues = append(tmpValues, fmt.Sprintf("(%s)", strings.Join(values[keysLen-len(keys):keysLen], ",")))
			}
			keysLen++
		}
		insertValue = strings.Join(tmpValues, ",")
	} else {
		insertValue = fmt.Sprintf("(%s)", strings.Join(values, ","))
	}
	query := fmt.Sprintf("insert into %s (%s) values %s", q.table, strings.Join(keys, ","), insertValue)
	log.Printf("insert sql:%s\n", query)
	state, err := q.db.Prepare(query)
	if err != nil {
		return 0, err
	}
	result, err := state.Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
