package common

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

type Data struct {
	D1 Duration `json:"d1"`
}

func TestEncode(t *testing.T) {

	data := Data{
		D1: ParseDurationOrPanic("25s"),
	}
	builder := &strings.Builder{}
	err := json.NewEncoder(builder).Encode(data)
	if err != nil {
		panic(err)
	}
	result := builder.String()
	//fmt.Printf(result)
	assert.Contains(t, result, "25s")

}


func TestDecode(t *testing.T) {

	thejson := `{"d1":"30m0s"}`
	var data Data

	err := json.NewDecoder(strings.NewReader(thejson)).Decode(&data)
	if err != nil {
		panic(err)
	}
	expected, _ := time.ParseDuration("30m")

	assert.Equal(t, data.D1.Duration, expected)

}