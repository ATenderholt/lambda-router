package database

import (
	"fmt"
	"strings"
)

func buildDebug(query string, args ...interface{}) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "query [%s], ", query)
	builder.WriteString("parameters [")
	for arg := range args {
		fmt.Fprintf(&builder, "%v", arg)
	}
	builder.WriteString("]")

	return builder.String()
}
