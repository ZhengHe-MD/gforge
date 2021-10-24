package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"

	"github.com/ZhengHe-MD/gforge/dao"
	"github.com/ZhengHe-MD/gforge/schema"
	"github.com/didi/gendry/manager"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mkideal/cli"
)

func main() {
	if err := cli.Root(rootCommand, cli.Tree(help), cli.Tree(schemaCommand), cli.Tree(daoCommand)).Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

var help = cli.HelpCommand("display help information")

const version = "0.0.1"

type rootArg struct {
	cli.Helper
	Version bool `cli:"v" usage:"version"`
}

var rootCommand = &cli.Command{
	Name: "tools",
	Desc: "A collection of tools to generate code for operating database supported by Gendry",
	Argv: func() interface{} { return new(rootArg) },
	Fn: func(ctx *cli.Context) error {
		arg := ctx.Argv().(*rootArg)
		if arg.Version {
			ctx.String("%s\n", version)
			return nil
		}
		return nil
	},
}

type schemaArg struct {
	DBName    string `cli:"d" usage:"database name"`
	TableName string `cli:"t" usage:"table name"`
	UserName  string `cli:"u" usage:"user name"`
	Password  string `cli:"p" usage:"password"`
	Host      string `cli:"h" usage:"host" dft:"localhost"`
	Port      int    `cli:"P" usage:"port" dft:"3306"`
}

var schemaCommand = &cli.Command{
	Name: "table",
	Desc: "schema could generate go struct code for given table",
	Argv: func() interface{} { return new(schemaArg) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*schemaArg)
		_, err := getSchema(os.Stdout, argv)
		return err
	},
}

func getSchema(w io.Writer, argv *schemaArg) (string, error) {
	db, err := getDBInstance(&dBConfig{
		Host:     argv.Host,
		User:     argv.UserName,
		Password: argv.Password,
		Port:     argv.Port,
		DBName:   "information_schema",
	})
	if nil != err {
		return "", err
	}
	return schema.GetSchema(w, db, argv.TableName, argv.DBName)
}

var daoCommand = &cli.Command{
	Name: "dao",
	Desc: "dao generates code of dao layer by given table name",
	Argv: func() interface{} { return new(schemaArg) },
	Fn: func(ctx *cli.Context) (err error) {
		arg := ctx.Argv().(*schemaArg)
		var buff bytes.Buffer
		if _, err = io.Copy(&buff, addImport(arg.TableName)); err != nil {
			return
		}
		structName, err := getSchema(&buff, arg)
		if nil != err {
			return err
		}
		r, err := dao.GenerateDao(arg.TableName, structName)
		if nil != err {
			return err
		}
		_, err = io.Copy(&buff, r)
		if nil != err {
			return err
		}
		_, err = io.Copy(os.Stdout, &buff)
		return
	},
}

//DBConfig holds the basic configuration of database
type dBConfig struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Port     int    `json:"port"`
	DBName   string `json:"database"`
}

func getDBInstance(conf *dBConfig) (*sql.DB, error) {
	option := manager.New(conf.DBName, conf.User, conf.Password, conf.Host)
	return option.Port(conf.Port).Open(true)
}

func addImport(packageName string) io.Reader {
	return bytes.NewBuffer([]byte(fmt.Sprintf(`
	package %s
	import (
		"database/sql"
		"errors"
		"github.com/didi/gendry/builder"
		"github.com/didi/gendry/scanner"
	)

	/*
	This code is generated by gendry
	*/
	`, packageName)))
}
