package dbinterface

import (
	"database/sql"
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
)

type TermDef struct {
	term       string
	definition string
}

func Connect(cfg mysql.Config, tableName string) (*DatabaseConn, error) {
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return &DatabaseConn{}, fmt.Errorf("Connect error: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return &DatabaseConn{}, fmt.Errorf("Connect error: %v", err)
	}
	fmt.Println("Connected!")

	databaseConn := &DatabaseConn{
		db:           db,
		tableName:    tableName,
		sequenceGaps: make([]int64, 0),
	}

	if err = databaseConn.checkForGaps(); err != nil {
		fmt.Printf("%v", err)
	}
	return databaseConn, nil
}

func addIfNotDuplicate(dbc *DatabaseConn, term string) (*int64, error) {
	foundId, err := dbc.findTerm(term)
	if err != nil {
		return nil, fmt.Errorf("Add %q: %v", term, err)
	}
	if len(foundId) != 0 {
		fmt.Printf("%q already exists at %v\n", term, foundId)
		return nil, nil
	}

	id, err := dbc.addTerm(term)
	if err != nil {
		return nil, fmt.Errorf("Add %q: %v", term, err)
	}
	return &id, nil
}

func Add(dbc *DatabaseConn, term string) ([]int64, error) {
	var addedIds []int64
	for _, c := range term {
		if !unicode.Is(unicode.Han, c) {
			return nil, fmt.Errorf("Add expecting Chinese characters, got %q", c)
		}
	}

	for _, c := range term {
		id, err := addIfNotDuplicate(dbc, string(c))
		if err != nil {
			return nil, err
		}
		if id != nil {
			addedIds = append(addedIds, *id)
		}
	}

	if utf8.RuneCountInString(term) > 1 {
		id, err := addIfNotDuplicate(dbc, term)
		if err != nil {
			return nil, err
		}
		if id != nil {
			addedIds = append(addedIds, *id)
		}
	}
	return addedIds, nil
}

func Delete(dbc *DatabaseConn, term string) error {
	err := dbc.deleteTerm(term)
	if err != nil {
		return fmt.Errorf("DeleteTerm: %v", err)
	}
	return nil
}

func Find(dbc *DatabaseConn, term string) (map[int64]TermDef, error) {
	terms, err := dbc.findAllTermsWithSubstring(term)
	if err != nil {
		return nil, fmt.Errorf("Find: %v", err)
	}
	return terms, nil
}

func List(dbc *DatabaseConn) (map[int64]TermDef, error) {
	listAll, err := dbc.listAll()
	if err != nil {
		return nil, fmt.Errorf("List: %v", err)
	}
	return listAll, nil
}
