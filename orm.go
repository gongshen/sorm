package sorm

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"
)

//实现orm
type Query struct {
	db     *sql.DB
	table  string
	wheres []string
	only   []string
	limit  string
	offset string
	order  string
	errs   []string
}

//select的
func (q *Query) toSQL() string {
	var where string
	if len(q.wheres) > 0 {
		where = fmt.Sprintf(`where %s`, q.wheres)
	}
	sqlStr := fmt.Sprintf(`select %s from %s %s %s %s %s`, strings.Join(q.only, ","), q.table, where, q.order, q.limit, q.offset)
	log.Printf(`select sql:%s`, sqlStr)
	return sqlStr
}

func (q *Query) setElem(rows *sql.Rows, t reflect.Type) (reflect.Value, error) {
	dest := reflect.New(t)
	addrs, err := address(dest, q.only)
	if err != nil {
		return reflect.ValueOf(nil), err
	}
	if len(q.only) != len(addrs) {
		return reflect.ValueOf(nil), errors.New("method setElem error: columns not match address!")
	}
	if err = rows.Scan(addrs...); err != nil {
		return reflect.ValueOf(nil), err
	}
	return dest, nil
}

//对基类型进行取地址操作
func address(dest reflect.Value, columns []string) ([]interface{}, error) {
	dest = dest.Elem()
	t := dest.Type()
	addr := make([]interface{}, 0)
	switch t.Kind() {
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			tf := t.Field(i)
			vf := dest.Field(i)
			if vf.Type().Name() == "Time" {
				return nil, errors.New("Method address: Type error,can't use Time!")
			}
			if tf.Anonymous {
				continue
			}
			if vf.Type().Kind() == reflect.Ptr {
				vf = vf.Elem()
			}
			//如果是struct类型并且字段类型不为Time，递归获取地址
			if vf.Kind() == reflect.Struct {
				ntf := reflect.New(vf.Type())
				vf.Set(ntf.Elem())
				caddrs, _ := address(vf, columns)
				addr = append(addr, caddrs...)
				continue
			}

			//取出tag字段
			column := strings.Split(tf.Tag.Get("json"), ",")[0]
			if column == "" {
				continue
			}
			//筛选出要求的字段
			for _, col := range columns {
				if column == col {
					addr = append(addr, vf.Addr().Interface())
					break
				}
			}
		}
	default:
		addr = append(addr, dest.Addr().Interface())
	}
	return addr, nil
}

func Connect(dsn string) (*sql.DB, error) {
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	//设置连接池
	conn.SetMaxOpenConns(100)
	conn.SetMaxIdleConns(10)
	conn.SetConnMaxLifetime(10 * time.Minute)
	return conn, conn.Ping()
}

//将数据库和表绑定
func Table(db *sql.DB, tableName string) func() *Query {
	return func() *Query {
		return &Query{
			db:    db,
			table: tableName,
		}
	}
}

//对struct类型获取键值对
func sKV(v reflect.Value) ([]string, []string) {
	var keys, values []string
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		tf := t.Field(i)
		vf := v.Field(i)
		//忽略非导出字段
		if tf.Anonymous {
			continue
		}
		//忽略无效、零值字段
		if !vf.IsValid() || reflect.DeepEqual(vf.Interface(), reflect.Zero(vf.Type())) {
			continue
		}
		//剥离指针
		for vf.Type().Kind() == reflect.Ptr {
			vf = vf.Elem()
		}
		//递归获取除time类型之外的组合struct
		if vf.Kind() == reflect.Struct && vf.Type().Name() != "Time" {
			cKeys, cValues := sKV(vf)
			keys = append(keys, cKeys...)
			values = append(values, cValues...)
			continue
		}
		//获取key，忽略无tag字段
		key := strings.Split(tf.Tag.Get("json"), ",")[0]
		if key == "" {
			continue
		}
		value := format(vf)
		if value != "" {
			values = append(values, value)
			keys = append(keys, key)
		}
	}
	return keys, values
}

//对map进行取键值对
func mKV(v reflect.Value) ([]string, []string) {
	var keys, values []string
	mapKeys := v.MapKeys()
	for _, key := range mapKeys {
		value := format(v.MapIndex(key))
		values = append(values, value)
		keys = append(keys, key.Interface().(string))
	}
	return keys, values
}

func format(v reflect.Value) string {
	//如果是time.Time类型的直接转unix时间戳
	if t, ok := v.Interface().(time.Time); ok {
		return fmt.Sprintf("From_UnixTime(%d)", t.Unix())
	}
	switch v.Kind() {
	case reflect.String:
		return fmt.Sprintf("'%s'", v.Interface())
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return fmt.Sprintf("%d", v.Interface())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", v.Interface())
	case reflect.Slice:
		var values []string
		for i := 0; i < v.Len(); i++ {
			//递归格式化
			values = append(values, format(v.Index(i)))
			return fmt.Sprintf("%s", strings.Join(values, ","))
		}
	//接口类型再剥调一层
	case reflect.Interface:
		return format(v.Elem())
	default:
		return ""
	}
	return ""
}

//User,*User,map[string]interface{}
func (q *Query) Where(wheres ...interface{}) *Query {
	for _, w := range wheres {
		str, err := Where(true, w)
		q.wheres = append(q.wheres, str)
		if err != nil {
			q.errs = append(q.errs, err.Error())
		}
	}
	return q
}

//where的逻辑处理
func Where(eq bool, in interface{}) (string, error) {
	var keys, values []string
	v := reflect.ValueOf(in)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.String:
		//return v.Interface().(string), nil
		return in.(string), nil
	case reflect.Struct:
		keys, values = sKV(v)
	case reflect.Map:
		keys, values = mKV(v)
	default:
		return "", errors.New("method Where error:type error!")
	}
	//判断键值对是否匹配
	if len(keys) != len(values) {
		return "", errors.New("method Where error:unmatch key and value!")
	}
	var wheres []string
	for idx, key := range keys {
		if eq {
			if strings.HasPrefix(values[idx], "(") && strings.HasSuffix(values[idx], ")") {
				wheres = append(wheres, fmt.Sprintf("%s in %s", key, values[idx]))
				continue
			}
			wheres = append(wheres, fmt.Sprintf("%s = %s"), key, values[idx])
			continue
		}
		if strings.HasPrefix(values[idx], ")") && strings.HasSuffix(values[idx], ")") {
			wheres = append(wheres, fmt.Sprintf("%s not in %s", key, values[idx]))
			continue
		}
		wheres = append(wheres, fmt.Sprintf("%s != %s"), key, values[idx])
	}
	return strings.Join(wheres, " and "), nil
}

func (q *Query) WhereNot(wheres ...interface{}) *Query {
	for _, w := range wheres {
		str, err := Where(false, w)
		q.wheres = append(q.wheres, str)
		if err != nil {
			q.errs = append(q.errs, err.Error())
		}
	}
	return q
}

//Limit(4)
func (q *Query) Limit(limit uint) *Query {
	q.limit = fmt.Sprintf("limit %d", limit)
	return q
}

//Offset(9)
//limit 4 offset 9：表示显示10，11，12，13页
func (q *Query) Offset(offset uint) *Query {
	q.offset = fmt.Sprintf("offset %d", offset)
	return q
}

//Order()
func (q *Query) Order(ord string) *Query {
	q.order = fmt.Sprintf("order by %s", ord)
	return q
}

//Only("first_name","last_name")
func (q *Query) Only(columns ...string) *Query {
	q.only = append(q.only, columns...)
	return q
}

//只获取struct的key
func sK(v reflect.Value) []string {
	var keys []string
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		tf := t.Field(i)
		vf := v.Field(i)
		//忽略非导出字段
		if tf.Anonymous {
			continue
		}
		//剥离指针
		for vf.Type().Kind() == reflect.Ptr {
			vf = vf.Elem()
		}
		//递归获取keys
		if vf.Kind() == reflect.Struct && vf.Type().Name() != "Time" {
			keys = append(keys, sK(vf)...)
			continue
		}
		//获取key，忽略无tag的字段
		key := strings.Split(tf.Tag.Get("json"), ",")[0]
		if key == "" {
			continue
		}
		keys = append(keys, key)
	}
	return keys
}
