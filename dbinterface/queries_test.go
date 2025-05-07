package dbinterface

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/flashcards/database"
	"github.com/go-sql-driver/mysql"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var dbc *DatabaseConn

func createDbContainer(ctx context.Context, databaseName string) (testcontainers.Container, string, error) {
	port := "3306"

	env := map[string]string{
		"MYSQL_ROOT_PASSWORD": "secret",
		"MYSQL_DATABASE":      databaseName,
		"MYSQL_USER":          databaseName,
		"MYSQL_PASSWORD":      "secret",
	}

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mysql:8.4.4",
			ExposedPorts: []string{port},
			Env:          env,
			Name:         databaseName,
			WaitingFor:   wait.ForLog("port: 3306  MySQL Community Server - GPL"),
		},
		Started: true,
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, "", err
	}

	mappedPort, err := container.MappedPort(ctx, nat.Port(port))
	if err != nil {
		return nil, "", err
	}

	return container, mappedPort.Port(), nil
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	databaseName := "test-db"
	container, port, err := createDbContainer(ctx, databaseName)
	if err != nil {
		log.Fatal(err)
	}

	tableName := "term"

	if err := database.CreateTable(ctx, port, databaseName, "secret", tableName); err != nil {
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
	dbc, err = Connect(cfg, tableName)
	if err != nil {
		log.Fatal(err)
	}

	_, err = Add(dbc, "我")
	if err != nil {
		log.Fatal(err)
	}

	m.Run()

	container.Terminate(ctx)

	os.Exit(0)
}

func TestAdd(t *testing.T) {
	type args struct {
		termToAdd string
		wantResp  []int64
		wantErr   bool
	}
	tests := map[string]args{
		"success": {
			termToAdd: "你",
			wantResp:  []int64{2},
			wantErr:   false,
		},
		"adding duplicate": {
			termToAdd: "我",
			wantResp:  nil,
			wantErr:   false,
		},
		"not Chinese character": {
			termToAdd: "c",
			wantResp:  nil,
			wantErr:   true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := Add(dbc, test.termToAdd)
			if err != nil && test.wantErr == false {
				t.Errorf("Got error %q, wanted nil", err.Error())
			}
			if !reflect.DeepEqual(got, test.wantResp) {
				t.Errorf("Got %v; wanted %v", got, test.wantResp)
			}
		})
	}
}

func TestFindTerm(t *testing.T) {
	type args struct {
		termToFind string
		wantResp   []int64
		wantErr    error
	}
	tests := map[string]args{
		"success": {
			termToFind: "我",
			wantResp:   []int64{1},
			wantErr:    nil,
		},
		"not found": {
			termToFind: "她",
			wantErr:    nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := dbc.findTerm(test.termToFind)
			if err != test.wantErr {
				t.Errorf("Got error %v; wanted %v", err, test.wantErr)
			}
			if !reflect.DeepEqual(got, test.wantResp) {
				t.Errorf("Got %v; wanted %v", got, test.wantResp)
			}
		})
	}
}
