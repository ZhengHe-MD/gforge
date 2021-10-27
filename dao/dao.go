package dao

import (
	"bytes"
	"go/format"
	"io"
	"text/template"
)

const (
	daoCode = `
//GetOne{{.StructName}} gets one record from table {{.TableName}} by condition "where"
func GetOne{{.StructName}}(db *sql.DB, where map[string]interface{}) (*domain.{{.StructName}}, error) {
	if nil == db {
		return nil, errors.New("sql.DB object couldn't be nil")
	}
	cond,vals,err := builder.BuildSelect("{{.TableName}}", where, nil)
	if nil != err {
		return nil, err
	}
	row,err := db.Query(cond, vals...)
	if nil != err || nil == row {
		return nil, err
	}
	defer row.Close()
	var res *domain.{{.StructName}}
	err = scanner.Scan(row, &res)
	return res,err
}

//GetMulti{{.StructName}} gets multiple records from table {{.TableName}} by condition "where"
func GetMulti{{.StructName}}(db *sql.DB, where map[string]interface{}) ([]*domain.{{.StructName}}, error) {
	if nil == db {
		return nil, errors.New("sql.DB object couldn't be nil")
	}
	cond,vals,err := builder.BuildSelect("{{.TableName}}", where, nil)
	if nil != err {
		return nil, err
	}
	row,err := db.Query(cond, vals...)
	if nil != err || nil == row {
		return nil, err
	}
	defer row.Close()
	var res []*domain.{{.StructName}}
	err = scanner.Scan(row, &res)
	return res,err
}

//Insert{{.StructName}} inserts an array of data into table {{.TableName}}
func Insert{{.StructName}}(db *sql.DB, data []map[string]interface{}) (int64, error) {
	if nil == db {
		return 0, errors.New("sql.DB object couldn't be nil")
	}
	cond, vals, err := builder.BuildInsert("{{.TableName}}", data)
	if nil != err {
		return 0, err
	}
	result,err := db.Exec(cond, vals...)
	if nil != err || nil == result {
		return 0, err
	}
	return result.LastInsertId()
}

//Update{{.StructName}} updates the table {{.TableName}}
func Update{{.StructName}}(db *sql.DB, where,data map[string]interface{}) (int64, error) {
	if nil == db {
		return 0, errors.New("sql.DB object couldn't be nil")
	}
	cond,vals,err := builder.BuildUpdate("{{.TableName}}", where, data)
	if nil != err {
		return 0, err
	}
	result,err := db.Exec(cond, vals...)
	if nil != err {
		return 0, err
	}
	return result.RowsAffected()
}

// Delete deletes matched records in {{.TableName}}
func Delete{{.StructName}}(db *sql.DB, where,data map[string]interface{}) (int64, error) {
	if nil == db {
		return 0, errors.New("sql.DB object couldn't be nil")
	}
	cond,vals,err := builder.BuildDelete("{{.TableName}}", where)
	if nil != err {
		return 0, err
	}
	result,err := db.Exec(cond, vals...)
	if nil != err {
		return 0, err
	}
	return result.RowsAffected()
}
`
)

type fillData struct {
	StructName string
	TableName  string
}

// GenerateDao generates Dao code
func GenerateDao(tableName, structName string) (io.Reader, error) {
	var buff bytes.Buffer
	err := template.Must(template.New("dao").Parse(daoCode)).Execute(&buff, fillData{
		StructName: structName,
		TableName:  tableName,
	})
	if err != nil {
		return nil, err
	}

	formatted, err := format.Source(buff.Bytes())
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(formatted), nil
}
