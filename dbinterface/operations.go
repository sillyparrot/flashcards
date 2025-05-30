package dbinterface

import (
	"database/sql"
	"errors"
	"log"
	"unicode"
	"unicode/utf8"

	"github.com/flashcards/dict"
	"github.com/go-sql-driver/mysql"
)

type TermDef struct {
	Term       string
	Definition string
}

func Connect(cfg mysql.Config, tableName string) (*DatabaseConn, error) {
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return &DatabaseConn{}, err
	}

	err = db.Ping()
	if err != nil {
		return &DatabaseConn{}, err
	}
	log.Println("Connected!")

	databaseConn := &DatabaseConn{
		db:           db,
		tableName:    tableName,
		sequenceGaps: make([]int64, 0),
	}

	if err = databaseConn.checkForGaps(); err != nil {
		log.Printf("%v", err)
	}
	return databaseConn, nil
}

func verifyLanguage(term string) error {
	for _, c := range term {
		if !unicode.Is(unicode.Han, c) {
			return &ErrUnexpectedLanguage{expectedLanguage: "Chinese", term: string(c)}
		}
	}
	return nil
}

func addIfNotDuplicate(dbc *DatabaseConn, term string, dictMap dict.DictMap) (*int64, error) {
	foundId, err := dbc.findTerm(term)
	var notFound *ErrNotFound
	if !errors.As(err, &notFound) && err != nil {
		return nil, err
	}
	if len(foundId) != 0 {
		log.Printf("%q already exists at %v\n", term, foundId)
		return nil, nil
	}

	def, inDict := dictMap.GetDefinition(term)
	if !inDict {
		log.Printf("%q not found in dictionary", term)
	} else {
		log.Printf("%q found in dictionary, definition: %q", term, def)
	}

	id, err := dbc.addTerm(term, def)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func Add(dbc *DatabaseConn, term string, dictMap dict.DictMap) ([]int64, error) {
	var addedIds []int64
	err := verifyLanguage(term)
	if err != nil {
		return nil, err
	}

	for _, c := range term {
		id, err := addIfNotDuplicate(dbc, string(c), dictMap)
		if err != nil {
			return nil, err
		}
		if id != nil {
			addedIds = append(addedIds, *id)
		}
	}

	if utf8.RuneCountInString(term) > 1 {
		id, err := addIfNotDuplicate(dbc, term, dictMap)
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
	err := verifyLanguage(term)
	if err != nil {
		return err
	}

	err = dbc.deleteTerm(term)
	if err != nil {
		return err
	}
	return nil
}

func Find(dbc *DatabaseConn, term string) (map[int64]TermDef, error) {
	err := verifyLanguage(term)
	if err != nil {
		return nil, err
	}

	terms, err := dbc.findAllTermsWithSubstring(term)
	if err != nil {
		return nil, err
	}
	return terms, nil
}

func List(dbc *DatabaseConn) (map[int64]TermDef, error) {
	listAll, err := dbc.listAll()
	if err != nil {
		return nil, err
	}
	return listAll, nil
}
