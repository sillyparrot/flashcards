package dict

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type DictionaryEntry struct {
	Traditional string
	Simplified  string
	Pinyin      string
	English     string
}

var listOfEntries []DictionaryEntry

func parseLine(line string) {
	if line == "" || line[0] == '#' {
		return
	}
	line = strings.TrimRight(line, "/")
	parts := strings.Split(line, "/")
	if len(parts) <= 1 {
		return
	}

	english := parts[1]
	charAndPinyin := strings.Split(parts[0], "[")
	if len(charAndPinyin) < 2 {
		return
	}

	characters := strings.Fields(charAndPinyin[0])
	if len(characters) < 2 {
		return
	}
	traditional := characters[0]
	simplified := characters[1]

	pinyin := strings.TrimRight(charAndPinyin[1], "] \n")

	entry := DictionaryEntry{
		Traditional: traditional,
		Simplified:  simplified,
		Pinyin:      pinyin,
		English:     english,
	}
	listOfEntries = append(listOfEntries, entry)
}

func removeSurnames() {
	for i := len(listOfEntries) - 2; i >= 0; i-- {
		if strings.Contains(listOfEntries[i].English, "surname ") &&
			listOfEntries[i].Traditional == listOfEntries[i+1].Traditional {
			listOfEntries = append(listOfEntries[:i], listOfEntries[i+1:]...)
		}
	}
}

func parse() {
	file, err := os.Open("cedict_ts.u8")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	fmt.Println("Parsing dictionary . . .")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parseLine(scanner.Text())
	}

	fmt.Println("Removing Surnames . . .")
	removeSurnames()

	fmt.Println("Done!")
	// Optionally: Do something with listOfDicts, like writing to JSON or a database.
}
