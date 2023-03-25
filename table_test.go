package sqlver_test

import (
	"context"
	"testing"

	"github.com/cbart/sqlver"
	"github.com/cbart/sqlver/sqlvertest"
	"github.com/google/go-cmp/cmp"
)

func TestCreateTable(t *testing.T) {
	t.Parallel()
	db := sqlvertest.DB(t)
	table := sqlver.Table{
		Name: "name",
		Columns: []sqlver.Column{
			sqlver.Integer("i"),
			sqlver.Text("t"),
			sqlver.Boolean("b"),
		},
	}
	if err := table.Up(context.Background(), db); err != nil {
		t.Errorf("failed to up table: %s", err)
	}
	rows, err := db.Query(`
		SELECT 
			column_name, 
			data_type 
		FROM 
			information_schema.columns
		WHERE 
			table_name = $1
	`, table.Name)
	if err != nil {
		t.Fatalf("failed to query schema of %q: %s", table.Name, err)
	}
	defer rows.Close()
	want := map[string]string{
		"i": "integer",
		"t": "text",
		"b": "boolean",
	}
	got := map[string]string{}
	for rows.Next() {
		var columnName, dataType string
		if err := rows.Scan(&columnName, &dataType); err != nil {
			t.Fatalf("cannot scan schema row: %s", err)
		}
		got[columnName] = dataType
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("table schema (column name -> data type) -want+got: %s", diff)
	}
}
