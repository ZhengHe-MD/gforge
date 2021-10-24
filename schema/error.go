package schema

import (
	"fmt"
)

const (
	errFormat = "Schema Error:[%s]\n"
)

func errUnknownType(columnName, columnType string) error {
	return schemaError(fmt.Sprintf("unknown datatype: columnName:%s, columnType:[%s]", columnName, columnType))
}

func schemaError(errmsg string) error {
	return fmt.Errorf(errFormat, errmsg)
}
