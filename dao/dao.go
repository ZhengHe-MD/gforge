package dao

import (
	"bytes"
	"go/format"
	"io"
	"text/template"
)

const (
	daoCode = `
//GetOne{{.StructName}} gets one record from table {{.TableName}} by "where".
func GetOne{{.StructName}}(ctx context.Context, db *sql.DB, where map[string]interface{}) (*domain.{{.StructName}}, error) {
	query, args, err := builder.BuildSelect("{{.TableName}}", where, nil)
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{"query": query, "args": args}).Debugln("GetOne{{.StructName}}")
	row, err := db.QueryContext(ctx, query, args...)
	if err != nil || row == nil {
		return nil, err
	}
	defer row.Close()
	var res *domain.{{.StructName}}
	err = scanner.Scan(row, &res)
	return res,err
}

//GetMulti{{.StructName}} gets multiple records from table {{.TableName}} by "where".
func GetMulti{{.StructName}}(ctx context.Context, db *sql.DB, where map[string]interface{}) ([]*domain.{{.StructName}}, error) {
	query,args,err := builder.BuildSelect("{{.TableName}}", where, nil)
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{"query": query, "args": args}).Debugln("GetMulti{{.StructName}}")
	row, err := db.QueryContext(ctx, query, args...)
	if err != nil || row == nil {
		return nil, err
	}
	defer row.Close()
	var res []*domain.{{.StructName}}
	err = scanner.Scan(row, &res)
	return res,err
}

//Insert{{.StructName}} inserts an array of data into table {{.TableName}}.
func Insert{{.StructName}}(ctx context.Context, db *sql.DB, data []map[string]interface{}) (int64, error) {
	query, args, err := builder.BuildInsert("{{.TableName}}", data)
	if err != nil {
		return 0, err
	}
	log.WithFields(log.Fields{"query": query, "args": args}).Debugln("Insert{{.StructName}}")
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil || result == nil {
		return 0, err
	}
	return result.LastInsertId()
}

//Update{{.StructName}} updates the table {{.TableName}}.
func Update{{.StructName}}(ctx context.Context, db *sql.DB, where,data map[string]interface{}) (int64, error) {
	query, args, err := builder.BuildUpdate("{{.TableName}}", where, data)
	if err != nil {
		return 0, err
	}
	log.WithFields(log.Fields{"query": query, "args": args}).Debugln("Update{{.StructName}}")
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Delete deletes matched records in {{.TableName}}.
func Delete{{.StructName}}(ctx context.Context, db *sql.DB, where,data map[string]interface{}) (int64, error) {
	query, args, err := builder.BuildDelete("{{.TableName}}", where)
	if err != nil {
		return 0, err
	}
	log.WithFields(log.Fields{"query": query, "args": args}).Debugln("Delete{{.StructName}}")
	result,err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Upsert insert an record if it's not there, or update existing record otherwise, and returns the record id.
func Upsert{{.StructName}}(ctx context.Context, db *sql.DB, where, data map[string]interface{}) (int64, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			if e := tx.Rollback(); e != nil {
				log.WithField("err", e).Errorln("rollback upsert")
			}
			return
		}
		if e := tx.Commit(); e != nil {
			log.WithField("err", e).Errorln("commit upsert")
		}
	}()
	
	var query string
	var args []interface{}
	var result sql.Result
	var prev domain.{{.StructName}}
	query, args, err = builder.BuildSelect("{{.TableName}}", where, nil)
	if err != nil {
		return 0, err
	}
	log.WithFields(log.Fields{"query": query, "args": args}).Debugln("Upsert{{.StructName}}: select")
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil || rows == nil {
		return 0, err
	}
	err = scanner.Scan(rows, &prev)
	if err == scanner.ErrEmptyResult {
		query, args, err = builder.BuildInsert("{{.TableName}}", []map[string]interface{}{data})
		if err != nil {
			return 0, err
		}
		log.WithFields(log.Fields{"query": query, "args": args}).Debugln("Upsert{{.StructName}}: insert")
		result, err = tx.ExecContext(ctx, query, args...)
		if err != nil || result == nil {
			return 0, err
		}
		return result.LastInsertId()
	}
	query, args, err = builder.BuildUpdate("{{.TableName}}", where, data)
	if err != nil {
		return 0, err
	}
	log.WithFields(log.Fields{"query": query, "args": args}).Debugln("Upsert{{.StructName}}: update")
	result, err = tx.ExecContext(ctx, query, args...)
	if err != nil || result == nil {
		return 0, err
	}
	_, err = result.RowsAffected()
	return prev.Id, err
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
