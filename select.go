//查询数据库
package sorm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

//Select() 需要传指针，不能仅仅声明一个指针变量：var user *User，需要传地址
func (q *Query) Select(in interface{}) error {
	if len(q.errs) != 0 {
		return errors.New(strings.Join(q.errs, " "))
	}
	t := reflect.TypeOf(in)
	v := reflect.ValueOf(in)
	//判断in是否为指针类型
	if t.Kind() != reflect.Ptr {
		return errors.New("Method Select: Type error！")
	}
	if !v.Elem().CanAddr() {
		return errors.New("Method Select: Type error！")
	}
	//剥离指针
	t = t.Elem()
	v = v.Elem()
	//进行only的填充
	if len(q.only) == 0 {
		switch t.Kind() {
		case reflect.Struct:
			if t.Name() != "Time" {
				q.only = sK(v)
			}
		case reflect.Slice:
			t = t.Elem()
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			if t.Kind() == reflect.Struct {
				if t.Name() != "Time" {
					q.only = sK(reflect.Zero(t))
				}
			}
		}
	}
	if len(q.only) == 0 {
		return errors.New("Method Select:No columns to select！")
	}
	rows, err := q.db.Query(q.toSQL())
	if err != nil {
		return err
	}
	switch t.Kind() {
	case reflect.Map:
		return errors.New("Method Select: Type error")
	case reflect.Slice:
		df := t.Elem()
		for df.Kind() == reflect.Ptr {
			df = df.Elem()
		}
		if df.Kind() == reflect.Map {
			return errors.New("Method Select:Type error")
		}
		sl := reflect.MakeSlice(t, 0, 0)
		fmt.Println("查询结果：")
		for rows.Next() {
			var destination reflect.Value
			destination, err = q.setElem(rows, df)
			if err != nil {
				return err
			}
			sl = reflect.Append(sl, destination.Elem())

		}
		v.Set(sl)
		for i := 0; i < v.Len(); i++ {
			fmt.Println(v.Index(i))
		}
		return nil
	default:
		fmt.Println("查询结果：")
		for rows.Next() {
			destination, err := q.setElem(rows, t)
			if err != nil {
				return err
			}
			v.Set(destination.Elem())
			fmt.Println(v)
		}
	}
	return nil
}
