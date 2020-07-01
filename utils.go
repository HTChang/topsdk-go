package top

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

var (
	// China Standard Time
	cst = time.FixedZone("CST", 3600<<3) // 3600 * 8
)

func interfaceToString(src interface{}) string {
	if src == nil {
		log.Fatal("src should not be nil")
	}
	switch src.(type) {
	case string:
		return src.(string)
	case int, int8, int32, int64:
	case uint8, uint16, uint32, uint64:
	case float32, float64:
		return fmt.Sprint(src)
	}
	data, err := json.Marshal(src)
	if err != nil {
		log.Fatal(err)
	}
	return string(data)
}
