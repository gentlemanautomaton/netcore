package schema

import "strings"

// clean removes any invalid characters from a provided key, including the
// slash '/' path separator.
func clean(key string) string {
	return strings.Replace(key, "/", "", -1)
}
