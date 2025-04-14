package main

/* Flashcards app
Users will input words they want to make into flashcards.
The app will check the database to prevent duplicates.
If there are multiple words, it will check individual words
and the entire phrase. For each new word / phrase, it will
connect to a dictionary and get the dictionary definition.
Users will be able to edit the definition afterwards.
It will save the words and definition in the database.
For words with multiple pronounciations and meanings it will
combine them all into one card.
The app will also test the user, it will shuffle the flashcards
and show them to the user one at a time. The user can choose if
they want to be shown the word or the definition. The user can
click a button to reveal the other side of the flashcard and
select three buttons ("got it right", "unsure", "got it wrong").
The app will store the response in the database as well. Later,
the user can choose to focus on "unsure", "got it wrong", and
new cards.
*/

import (
	"fmt"
	"log"
	"os"
	"sort"

	dbinterface "github.com/flashcards/database_interface"
	"github.com/go-sql-driver/mysql"
)

func add(dbc *dbinterface.DatabaseConn) {
	fmt.Println("Enter the term(s) you want to add to the database. Type menu to return to menu.")
	for {
		var input string
		_, err := fmt.Scan(&input)
		if err != nil {
			log.Fatalf("input error %v", err)
		}
		if input == "menu" {
			return
		}
		ids, err := dbinterface.Add(dbc, input)
		if err != nil {
			log.Fatal(err)
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
			log.Fatalf("input error %v", err)
		}
		if input == "menu" {
			return
		}
		err = dbinterface.Delete(dbc, input)
		if err != nil {
			log.Fatal(err)
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
			log.Fatalf("input error %v", err)
		}
		if input == "menu" {
			return
		}
		terms, err := dbinterface.Find(dbc, input)
		if err != nil {
			log.Fatal(err)
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
		log.Fatal(err)
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
