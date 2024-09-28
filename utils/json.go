/*
@author: sk
@date: 2024/9/17
*/
package utils

import (
	"encoding/json"
	"fmt"
)

func ToString(obj any) string {
	bs, err := json.Marshal(obj)
	if err != nil {
		fmt.Printf("json marshal err %v\n", err)
		return ""
	}
	return string(bs)
}
