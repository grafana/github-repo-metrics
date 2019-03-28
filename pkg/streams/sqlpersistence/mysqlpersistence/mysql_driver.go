package mysqlpersistence

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"

	// make sure to load mysql driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/grafana/devtools/pkg/streams"
	"github.com/grafana/devtools/pkg/streams/sqlpersistence"
)

func init() {
	sqlpersistence.Register("mysql", &mySqlDriver{})
}

type mySqlDriver struct {
}

func (sp *mySqlDriver) DropTableIfExists(tx *sql.Tx, t *sqlpersistence.Table) error {
	_, err := tx.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS %s`, t.TableName))
	if err != nil {
		return err
	}

	return nil
}

func (sp *mySqlDriver) CreateTableIfNotExists(tx *sql.Tx, t *sqlpersistence.Table) error {
	var createTable bytes.Buffer
	createTable.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s ( ", t.TableName))
	primaryKeys := []string{}
	for _, c := range t.Columns {
		createTable.WriteString(c.ColumnName + " ")
		createTable.WriteString(c.ColumnType + " ")
		createTable.WriteString("NOT NULL, ")
		if c.PrimaryKey {
			primaryKeys = append(primaryKeys, c.ColumnName)
		}
	}
	createTable.WriteString("PRIMARY KEY(")
	createTable.WriteString(strings.Join(primaryKeys, ","))
	createTable.WriteString("))")
	_, err := tx.Exec(createTable.String())

	if err != nil {
		return err
	}

	return nil
}

func (sp *mySqlDriver) PersistStream(tx *sql.Tx, t *sqlpersistence.Table, stream streams.Readable) error {
	buf := &bytes.Buffer{}
	buf.WriteString("INSERT INTO ")
	buf.WriteString(t.TableName)
	buf.WriteString(" (")
	buf.WriteString(strings.Join(t.GetColumnNames(), ","))
	buf.WriteString(") VALUES ")
	initialSQL := buf.String()

	values := []interface{}{}

	preparedArgs := []string{}
	for n := 0; n < len(t.GetColumnNames()); n++ {
		preparedArgs = append(preparedArgs, "?")
	}
	preparedSQLStr := "(" + strings.Join(preparedArgs, ",") + ")"
	sql := ""
	processedRows := int64(0)

	for msg := range stream {
		colValues := t.GetColumnValues(msg)
		if len(colValues) == 0 {
			continue
		}
		values = append(values, colValues...)

		if processedRows != 0 {
			sql += ","
		}

		sql += preparedSQLStr
		processedRows++

		if processedRows > 999 {
			stmt, err := tx.Prepare(initialSQL + sql)
			if err != nil {
				return err
			}

			_, err = stmt.Exec(values...)
			if err != nil {
				return err
			}

			err = stmt.Close()
			if err != nil {
				return err
			}

			processedRows = 0
			sql = ""
			values = []interface{}{}
		}
	}

	if len(values) == 0 {
		return nil
	}

	stmt, err := tx.Prepare(initialSQL + sql)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(values...)
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	return nil
}