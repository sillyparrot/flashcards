package main

import (
	"fmt"
	"slices"
)

func append_list(list *[]string) {
	*list = append(*list, "a")
}

func delete_list(list []string) []string {
	return slices.Delete(list, 2, 3)
}

func main() {
	list := []string{"a", "b", "c"}
	append_list(&list)
	list = delete_list(list)
	fmt.Println(list)
}
