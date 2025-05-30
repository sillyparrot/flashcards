package dict

type DictionaryEntry struct {
	Traditional string
	Simplified  string
	Pinyin      string
	English     string
}

type DictMap map[string]DictionaryEntry

func (d *DictMap) GetDefinition(term string) (string, bool) {
	entry, ok := (*d)[term]
	if !ok {
		return "", false
	}
	return entry.English, ok
}
