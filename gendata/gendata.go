package gendata

import (
	"bytes"
	"fmt"
	"github.com/yuin/gopher-lua"
	"go-randgen/gendata/generators"
	"math/rand"
	"strconv"
	"strings"
)


type ZzConfig struct {
	Tables *Tables
	Fields *Fields
	Data   *Data
}

func newZzConfig(l *lua.LState) (*ZzConfig, error) {
	tables, err := newTables(l)
	if err != nil {
		return nil, err
	}

	fields, err := newFields(l)
	if err != nil {
		return nil, err
	}

	data, err := newData(l)
	if err != nil {
		return nil, err
	}

	return &ZzConfig{Tables:tables, Fields:fields, Data:data}, nil
}

func (z *ZzConfig) genDdls() ([]*tableStmt, []*fieldExec, error) {
	tableStmts, err := z.Tables.gen()
	if err != nil {
		return nil, nil, err
	}

	fieldStmts, fieldExecs, err := z.Fields.gen()
	if err != nil {
		return nil, nil, err
	}

	for _, tableStmt := range tableStmts {
		tableStmt.wrapInTable(fieldStmts)
	}

	return tableStmts, fieldExecs, nil
}

func ByZz(zz string) ([]string, Keyfun, error) {
	l ,err := runLua(zz)
	if err != nil {
		return nil, nil, err
	}

	config, err := newZzConfig(l)
	if err != nil {
		return nil, nil, err
	}

	return  ByConfig(config)
}

func ByConfig(config *ZzConfig) ([]string, Keyfun, error)  {
	tableStmts, fieldExecs, err := config.genDdls()
	if err != nil {
		return nil, nil, err
	}

	recordGor := config.Data.getRecordGen(fieldExecs)
	row := make([]string, len(fieldExecs))

	sqls := make([]string, 0, len(tableStmts))
	for _, tableStmt := range tableStmts {
		sqls = append(sqls, tableStmt.ddl)
		valuesStmt := make([]string, 0, tableStmt.rowNum)
		for i := 0; i < tableStmt.rowNum; i++ {
			recordGor.oneRow(row)
			valuesStmt = append(valuesStmt, wrapInDml(strconv.Itoa(i), row))
		}
		sqls = append(sqls, wrapInInsert(tableStmt.name, valuesStmt))
	}

	return sqls, newKeyfun(tableStmts, fieldExecs), nil
}

const insertTemp = "insert into %s values %s"

func wrapInInsert(tableName string, valuesStmt []string) string {
	return fmt.Sprintf(insertTemp, tableName, strings.Join(valuesStmt, ","))
}

func wrapInDml(pk string, data []string) string {
	buf := &bytes.Buffer{}
	buf.WriteString("(" + pk)

	for _, d := range data {
		buf.WriteString(",")
		if d == "null" {
			buf.WriteString(d)
		} else {
			buf.WriteString("\"" + d + "\"")
		}
	}

	buf.WriteString(")")

	return buf.String()
}


type Keyfun map[string]func() string

func newKeyfun(tables []*tableStmt, fields []*fieldExec) Keyfun {
	m := map[string]func() string{
		"_table": func() string {
			return tables[rand.Intn(len(tables))].name
		},
		"_field": func() string {
			return "`" + fields[rand.Intn(len(fields))].name + "`"
		},
		"_int": func() string {
			return generators.Get("int").Gen()
		},
		"_tinyint": func() string {
			return generators.Get("tinyint").Gen()
		},
		"_digit": func() string {
			return generators.Get("digit").Gen()
		},
		"_letter": func() string {
			return generators.Get("letter").Gen()
		},
		"_date": func() string {
			return generators.Get("date").Gen()
		},
		"_year": func() string {
			return generators.Get("year").Gen()
		},
		"_time": func() string {
			return generators.Get("time").Gen()
		},
		"_datetime": func() string {
			return generators.Get("datetime").Gen()
		},
		"_timestamp": func() string {
			return generators.Get("timestamp").Gen()
		},
		"_english": func() string {
			return generators.Get("english").Gen()
		},
	}

	return Keyfun(m)
}

func (k Keyfun) Gen(key string) (string, bool)  {
	if kf, ok := k[key]; ok {
		return kf(), true
	}
	return "", false
}
