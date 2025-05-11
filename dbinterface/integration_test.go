package dbinterface

import (
	"context"
	"errors"
	"fmt"
	"log"
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
}

func TestAdd(t *testing.T) {
	type args struct {
		termToAdd string
		wantResp  []int64
		wantErr   any
		cleanup   func(string)
	}
	tests := map[string]args{
		"success": {
			termToAdd: "你",
			wantResp:  []int64{2},
			wantErr:   nil,
			cleanup: func(termToDelete string) {
				err := Delete(dbc, termToDelete)
				if err != nil {
					t.Fatalf("Error when doing cleanup and deleting %s", termToDelete)
				}
			},
		},
		"adding duplicate": {
			termToAdd: "我",
			wantResp:  nil,
			wantErr:   nil,
		},
		"unexpected language": {
			termToAdd: "c",
			wantResp:  nil,
			wantErr:   errUnexpectedLanguage{},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := Add(dbc, test.termToAdd)
			if test.wantErr != nil && !errors.As(err, &test.wantErr) {
				t.Errorf("Got error %v, wanted %v", err, test.wantErr)
			} else if test.wantErr == nil && err != nil {
				t.Errorf("Got error, wanted nil")
			}
			if !reflect.DeepEqual(got, test.wantResp) {
				t.Errorf("Got %v; wanted %v", got, test.wantResp)
			}
			if test.cleanup != nil {
				test.cleanup(test.termToAdd)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		setup        func(string)
		termToDelete string
		wantErr      any
	}
	tests := map[string]args{
		"success": {
			termToDelete: "你",
			setup: func(termToAdd string) {
				_, err := Add(dbc, termToAdd)
				if err != nil {
					t.Fatalf("Error when adding term %s", termToAdd)
				}
			},
			wantErr: nil,
		},
		"term not found in database": {
			termToDelete: "爱",
			wantErr:      errNotFound{},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.setup != nil {
				test.setup(test.termToDelete)
			}
			err := Delete(dbc, test.termToDelete)
			if test.wantErr != nil && !errors.As(err, &test.wantErr) {
				t.Errorf("Got error %v, wanted %v", err, test.wantErr)
			} else if test.wantErr == nil && err != nil {
				t.Errorf("Got error, wanted nil")
			}
		})
	}
}

func TestFind(t *testing.T) {
	type args struct {
		setup      func()
		termToFind string
		wantResp   map[int64]TermDef
		wantErr    any
		cleanup    func()
	}
	tests := map[string]args{
		"find term": {
			termToFind: "我",
			wantResp: map[int64]TermDef{
				1: {term: "我", definition: ""},
			},
			wantErr: nil,
		},
		"not found": {
			termToFind: "她",
			wantErr:    errNotFound{},
		},
		"find all terms with substring": {
			setup: func() {
				_, err := Add(dbc, "我们")
				if err != nil {
					t.Fatalf("Error when adding term %s", "我们")
				}
			},
			termToFind: "我",
			wantResp: map[int64]TermDef{
				1: {term: "我", definition: ""},
				3: {term: "我们", definition: ""},
			},
			wantErr: nil,
			cleanup: func() {
				err := Delete(dbc, "我们")
				if err != nil {
					t.Fatalf("Error when deleting term %s", "我们")
				}
				err = Delete(dbc, "们")
				if err != nil {
					t.Fatalf("Error when deleting term %s", "我们")
				}
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.setup != nil {
				test.setup()
			}
			got, err := Find(dbc, test.termToFind)
			if test.wantErr != nil && !errors.As(err, &test.wantErr) {
				t.Errorf("Got error %v, wanted %v", err, test.wantErr)
			} else if test.wantErr == nil && err != nil {
				t.Errorf("Got error, wanted nil")
			}
			if !reflect.DeepEqual(got, test.wantResp) {
				t.Errorf("Got %v; wanted %v", got, test.wantResp)
			}
			if test.cleanup != nil {
				test.cleanup()
			}
		})
	}
}

func TestList(t *testing.T) {
	type args struct {
		wantResp map[int64]TermDef
		wantErr  any
	}
	tests := map[string]args{
		"list": {
			wantResp: map[int64]TermDef{
				1: {term: "我", definition: ""},
			},
			wantErr: nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := List(dbc)
			if test.wantErr != nil && !errors.As(err, &test.wantErr) {
				t.Errorf("Got error %v, wanted %v", err, test.wantErr)
			} else if test.wantErr == nil && err != nil {
				t.Errorf("Got error, wanted nil")
			}
			if !reflect.DeepEqual(got, test.wantResp) {
				t.Errorf("Got %v; wanted %v", got, test.wantResp)
			}
		})
	}
}
