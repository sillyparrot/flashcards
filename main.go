package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/flashcards/dbinterface"
	"github.com/go-sql-driver/mysql"
)

func add(dbc *dbinterface.DatabaseConn) {
	fmt.Println("Enter the term(s) you want to add to the database. Type menu to return to menu.")
	for {
		var input string
		_, err := fmt.Scan(&input)
		if err != nil {
			log.Printf("input error %v", err)
		}
		if input == "menu" {
			return
		}
		ids, err := dbinterface.Add(dbc, input)
		if err != nil {
			log.Print(err)
		}
		if len(ids) > 0 {
			fmt.Printf("Added IDs: %v\n", ids)
		}
	}
}

func delete(dbc *dbinterface.DatabaseConn) {
	fmt.Println("Enter the term(s) you want to delete from the database. Type menu to return to menu.")
	for {
		var input string
		_, err := fmt.Scan(&input)
		if err != nil {
			log.Printf("input error %v", err)
		}
		if input == "menu" {
			return
		}
		err = dbinterface.Delete(dbc, input)
		if err != nil {
			log.Print(err)
		}
	}
}

func printTerms(terms map[int64]dbinterface.TermDef) {
	var ids []int
	for t := range terms {
		ids = append(ids, int(t))
	}
	sort.Ints(ids)

	for _, t := range ids {
		fmt.Printf("%d: %v\n", t, terms[int64(t)])
	}
}

func find(dbc *dbinterface.DatabaseConn) {
	fmt.Println("Enter the term(s) you want to find in the database. Type menu to return to menu.")
	for {
		var input string
		_, err := fmt.Scan(&input)
		if err != nil {
			log.Printf("input error %v", err)
		}
		if input == "menu" {
			return
		}
		terms, err := dbinterface.Find(dbc, input)
		if err != nil {
			log.Print(err)
		}
		if len(terms) == 0 {
			fmt.Printf("no terms found with %s\n", input)
		}
		printTerms(terms)
	}
}

func list(dbc *dbinterface.DatabaseConn) {
	terms, err := dbinterface.List(dbc)
	if err != nil {
		log.Print(err)
	}
	if len(terms) == 0 {
		fmt.Println("No terms in flashcards database.")
		return
	}
	printTerms(terms)
}

func main() {
	cfg := mysql.Config{
		User:   os.Getenv("DBUSER"),
		Passwd: os.Getenv("DBPASS"),
		Net:    "tcp",
		Addr:   "127.0.0.1:3306",
		DBName: "flashcards",
	}
	tableName := "terms"
	dbc, err := dbinterface.Connect(cfg, tableName)
	if err != nil {
		log.Fatalf("%v", err)
	}

	for {
		fmt.Println("Select the operation you want to perform:")
		fmt.Println("1. Add")
		fmt.Println("2. Delete")
		fmt.Println("3. Find")
		fmt.Println("4. List")
		fmt.Println("Any other input: Exit")
		var input string
		_, err := fmt.Scan(&input)
		if err != nil {
			log.Fatalf("input error %v", err)
		}
		switch input {
		case "1":
			add(dbc)
		case "2":
			delete(dbc)
		case "3":
			find(dbc)
		case "4":
			list(dbc)
		default:
			return
		}
	}
}
