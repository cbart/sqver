package sqlvertest_test

import (
	"testing"

	"github.com/cbart/sqlver/sqlvertest"
)

func TestConnectToPostgres(t *testing.T) {
	sqlvertest.DB(t)
}
