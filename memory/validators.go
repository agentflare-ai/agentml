package memory

import (
	"fmt"
	"regexp"
)

var validTableName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func ValidateTableName(tableName string) error {
	if !validTableName.MatchString(tableName) {
		return fmt.Errorf("invalid table name: %s, must match %s", tableName, validTableName.String())
	}
	return nil
}
