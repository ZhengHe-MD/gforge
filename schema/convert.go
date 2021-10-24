package schema

import "strings"

func convertUnderScoreToCamel(name string) string {
	arr := strings.Split(name, cUnderScore)
	for i := 0; i < len(arr); i++ {
		arr[i] = lintName(strings.Title(arr[i]))
	}
	return strings.Join(arr, "")
}
