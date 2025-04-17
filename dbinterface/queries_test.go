package dbinterface

import (
	"fmt"
	"reflect"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestFindTerm(t *testing.T) {
	type args struct {
		termToFind string
		wantRows   *sqlmock.Rows
		want       []int64
		err        error
	}
	tests := map[string]args{
		"success": {
			termToFind: "我",
			wantRows:   sqlmock.NewRows([]string{"id"}).AddRow(1),
			want:       []int64{1},
			err:        nil,
		},
		"not found": {
			termToFind: "我",
			wantRows:   &sqlmock.Rows{},
			err:        nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				fmt.Println("failed to open sqlmock database:", err)
			}
			defer db.Close()

			dbc := DatabaseConn{
				db:           db,
				tableName:    "terms",
				sequenceGaps: make([]int64, 0),
			}

			mock.ExpectQuery("SELECT id").WillReturnRows(test.wantRows)

			got, err := dbc.findTerm(test.termToFind)
			if err != test.err {
				t.Errorf("Got error %v; wanted %v", err, test.err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Got %v; wanted %v", got, test.want)
			}
		})
	}
}
