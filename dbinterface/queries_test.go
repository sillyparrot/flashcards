package dbinterface

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/docker/go-connections/nat"
	"github.com/go-sql-driver/mysql"
	testcontainers "github.com/testcontainers/testcontainers-go"
)

func newDb(t *testing.T) (DatabaseConn, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}

	dbc := DatabaseConn{
		db:           db,
		tableName:    "terms",
		sequenceGaps: make([]int64, 0),
	}

	return dbc, mock, func() {
		db.Close()
	}
}

func createDbContainer(databaseName string) (testcontainers.Container, string, error) {
	port := "3306/tcp"

	var env = map[string]string{
		"DBUSER": databaseName,
		"DBPASS": "secret",
	}

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mysql:8.4.4",
			ExposedPorts: []string{port},
			Env:          env,
			Name:         databaseName,
		},
		Started: true,
	}

	container, err := testcontainers.GenericContainer(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}

	mappedPort, err := container.MappedPort(context.Background(), nat.Port(port))
	if err != nil {
		log.Fatal(err)
		return nil, "", err
	}

	return container, mappedPort.Port(), nil
}

func TestMain(m *testing.M) {
	databaseName := "test-db"
	container, port, err := createDbContainer(databaseName)
	if err != nil {
		log.Fatal(err)
	}
	addr := fmt.Sprintf("127.0.0.1:%s", port)
	cfg := mysql.Config{
		User:   databaseName,
		Passwd: "secret",
		Net:    "tcp",
		Addr:   addr,
		DBName: databaseName,
	}
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected!")
	defer db.Close()
	defer container.Terminate(context.Background())
}

func TestFindTerm(t *testing.T) {
	type args struct {
		termToFind string
		wantRows   *sqlmock.Rows
		want       []int64
		wantErr    error
	}
	tests := map[string]args{
		"success": {
			termToFind: "我",
			wantRows:   sqlmock.NewRows([]string{"id"}).AddRow(1),
			want:       []int64{1},
			wantErr:    nil,
		},
		"not found": {
			termToFind: "我",
			wantRows:   &sqlmock.Rows{},
			wantErr:    nil,
		},
	}
	for name, test := range tests {
		dbc, mock, teardown := newDb(t)
		defer teardown()
		t.Run(name, func(t *testing.T) {
			mock.ExpectQuery("SELECT id").WillReturnRows(test.wantRows)

			got, err := dbc.findTerm(test.termToFind)
			if err != test.wantErr {
				t.Errorf("Got error %v; wanted %v", err, test.wantErr)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Got %v; wanted %v", got, test.want)
			}
		})
	}
}
