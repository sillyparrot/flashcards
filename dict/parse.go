package dict

import (
	"bufio"
	"log"
	"os"
	"reflect"
	"slices"
	"strings"
)

func parseLine(line string) DictionaryEntry {
	if line == "" || line[0] == '#' {
		return DictionaryEntry{}
	}
	line = strings.TrimRight(line, "/")
	parts := strings.Split(line, "/")
	if len(parts) <= 1 {
		return DictionaryEntry{}
	}

	english := parts[1]
	charAndPinyin := strings.Split(parts[0], "[")
	if len(charAndPinyin) < 2 {
		return DictionaryEntry{}
	}

	characters := strings.Fields(charAndPinyin[0])
	if len(characters) < 2 {
		return DictionaryEntry{}
	}
	traditional := characters[0]
	simplified := characters[1]

	pinyin := strings.TrimRight(charAndPinyin[1], "] \n")

	return DictionaryEntry{
		Traditional: traditional,
		Simplified:  simplified,
		Pinyin:      pinyin,
		English:     english,
	}
}

func removeSurnames(listOfEntries []DictionaryEntry) []DictionaryEntry {
	for i := len(listOfEntries) - 2; i >= 0; i-- {
		if strings.Contains(listOfEntries[i].English, "surname ") &&
			listOfEntries[i].Traditional == listOfEntries[i+1].Traditional {
			listOfEntries = slices.Delete(listOfEntries, i, i+1)
		}
	}
	return listOfEntries
}

func ParseDict(filepath string) (DictMap, error) {
	var listOfEntries []DictionaryEntry

	file, err := os.Open(filepath)
	if err != nil {
		log.Println("Error opening file:", err)
		return nil, err
	}
	defer file.Close()

	log.Println("Parsing dictionary")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if entry := parseLine(scanner.Text()); !reflect.DeepEqual(entry, DictionaryEntry{}) {
			listOfEntries = append(listOfEntries, entry)
		}
	}

	listOfEntries = removeSurnames(listOfEntries)

	dictMap := make(DictMap)
	for _, entry := range listOfEntries {
		dictMap[entry.Simplified] = entry
	}
	return dictMap, nil
}
