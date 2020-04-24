package common

import (
	"encoding/json"
	"strings"
)

func JSON2String(data interface{}) string {
	builder := &strings.Builder{}
	_ = json.NewEncoder(builder).Encode(data)
	return builder.String()
}


