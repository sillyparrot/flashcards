package dbinterface

import (
	"database/sql"
	"fmt"
)

type DatabaseConn struct {
	db           *sql.DB
	tableName    string
	sequenceGaps []int64
}

func (dbc *DatabaseConn) findTerm(termToFind string) ([]int64, error) {
	query := fmt.Sprintf("SELECT id FROM %s WHERE term='%s'", dbc.tableName, termToFind)
	rows, err := dbc.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("findTerm %q: %v", termToFind, err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("findTerm %q: %v", termToFind, err)
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("findTerm %q: %v", termToFind, err)
	}

	if len(ids) == 0 {
		return nil, &errNotFound{term: termToFind}
	}

	return ids, nil
}

func (dbc *DatabaseConn) findAllTermsWithSubstring(termToFind string) (map[int64]TermDef, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE term LIKE '%%%s%%'", dbc.tableName, termToFind)
	rows, err := dbc.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("findAllTermsWithSubstring %q: %v", termToFind, err)
	}
	defer rows.Close()

	terms := make(map[int64]TermDef)
	for rows.Next() {
		var id int64
		var term string
		var def string
		if err := rows.Scan(&id, &term, &def); err != nil {
			return nil, fmt.Errorf("findAllTermsWithSubstring %q: %v", termToFind, err)
		}
		terms[id] = TermDef{term: term, definition: def}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("findAllTermsWithSubstring %q: %v", termToFind, err)
	}

	if len(terms) == 0 {
		return nil, &errNotFound{term: termToFind}
	}

	return terms, nil
}

func (dbc *DatabaseConn) checkForGaps() error {
	// if table is empty, return
	empty, err := dbc.tableIsEmpty()
	if err != nil {
		return err
	}
	if empty {
		return nil
	}

	gapQuery := fmt.Sprintf(`WITH RECURSIVE all_ids AS (
		SELECT MIN(id) AS id
		FROM %s
		UNION ALL
		SELECT id + 1
		FROM all_ids
		WHERE id + 1 <= (SELECT MAX(id) FROM %s)
	),
	existing_ids AS (
		SELECT id FROM %s
	)
	SELECT a.id AS missing_id
	FROM all_ids a
	LEFT JOIN existing_ids e ON a.id = e.id
	WHERE e.id IS NULL;
	`, dbc.tableName, dbc.tableName, dbc.tableName)
	rows, err := dbc.db.Query(gapQuery)
	if err != nil {
		return fmt.Errorf("checkForGaps: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var missingId int64
		if err := rows.Scan(&missingId); err != nil {
			return fmt.Errorf("checkForGaps: %v", err)
		}
		dbc.sequenceGaps = append(dbc.sequenceGaps, missingId)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("checkForGaps: %v", err)
	}
	if len(dbc.sequenceGaps) > 0 {
		fmt.Printf("Found gaps %v\n", dbc.sequenceGaps)
	}
	return nil
}

func (dbc *DatabaseConn) addTerm(term string) (int64, error) {
	var exec string
	var insertId int64
	if len(dbc.sequenceGaps) > 0 {
		fmt.Println("Inserting into sequence gap")
		insertId = dbc.sequenceGaps[0]
		exec = fmt.Sprintf("INSERT INTO %s (id, term, definition) VALUES (%d, '%s', '%s')", dbc.tableName, insertId, term, "")
	} else {
		exec = fmt.Sprintf("INSERT INTO %s (term, definition) VALUES ('%s', '%s')", dbc.tableName, term, "")
	}
	result, err := dbc.db.Exec(exec)
	if err != nil {
		return 0, fmt.Errorf("addTerm: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("addTerm: %v", err)
	}

	if insertId != 0 {
		if id != insertId {
			fmt.Printf("WARNING: inserted at %d, expected to insert at %d", id, insertId)
		} else {
			dbc.sequenceGaps = dbc.sequenceGaps[1:]
		}
	}

	fmt.Printf("Added %q\n", term)
	return id, nil
}

func (dbc *DatabaseConn) deleteTerm(term string) error {
	ids, err := dbc.findTerm(term)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		fmt.Printf("Term %q does not exist in database\n", term)
		return nil
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE term='%s'", dbc.tableName, term)
	result, err := dbc.db.Exec(query)
	if err != nil {
		return fmt.Errorf("deleteTerm: %v", err)
	}

	num, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("deleteTerm: %v", err)
	}
	if num != int64(len(ids)) {
		fmt.Printf("WARNING: %d number of rows deleted, expected %d", num, len(ids))
	} else {
		fmt.Printf("Deleted %q in rows %v\n", term, ids)
		dbc.sequenceGaps = append(dbc.sequenceGaps, ids...)
	}
	return nil
}

func (dbc *DatabaseConn) listAll() (map[int64]TermDef, error) {
	rows, err := dbc.db.Query(fmt.Sprintf("SELECT * from %s", dbc.tableName))
	if err != nil {
		return nil, fmt.Errorf("listAll: %v", err)
	}
	defer rows.Close()

	allTerms := make(map[int64]TermDef)
	for rows.Next() {
		var id int64
		var term string
		var def string
		if err := rows.Scan(&id, &term, &def); err != nil {
			return nil, fmt.Errorf("listAll: %v", err)
		}
		allTerms[id] = TermDef{term: term, definition: def}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("listAll: %v", err)
	}
	return allTerms, nil
}

func (dbc *DatabaseConn) tableIsEmpty() (bool, error) {
	rows, err := dbc.db.Query(fmt.Sprintf("SELECT id FROM %s LIMIT 1", dbc.tableName))
	if err != nil {
		return true, err
	}
	defer rows.Close()

	empty := true
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return empty, err
		}
		empty = false
	}
	if err := rows.Err(); err != nil {
		return empty, fmt.Errorf("tableIsEmpty: %v", err)
	}
	return empty, nil
}
