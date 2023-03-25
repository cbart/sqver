package sqlver

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
)

type Table struct {
	Name    string
	Columns []Column
}

func (t Table) Up(ctx context.Context, db *sql.DB) error {
	var ddl bytes.Buffer
	fmt.Fprintf(&ddl, "CREATE TABLE %s (", t.Name)
	var notFirstColumn bool
	for _, c := range t.Columns {
		if notFirstColumn {
			fmt.Fprint(&ddl, ",")
		}
		fmt.Fprintln(&ddl)
		c.fprintDDL(&ddl)
		notFirstColumn = true
	}
	fmt.Fprint(&ddl, ")\n")
	_, err := db.ExecContext(ctx, ddl.String())
	if err != nil {
		return err
	}
	return nil
}

type Column struct {
	name     string
	dataType string
}

func Integer(name string) Column { return Column{name: name, dataType: "integer"} }
func Text(name string) Column    { return Column{name: name, dataType: "text"} }
func Boolean(name string) Column { return Column{name: name, dataType: "boolean"} }

func (c Column) fprintDDL(w io.Writer) {
	fmt.Fprintf(w, "%s %s", c.name, c.dataType)
}
