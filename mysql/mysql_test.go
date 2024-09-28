/*
@author: sk
@date: 2024/9/15
*/
package main

import (
	"fmt"
	"testing"
)

func TestDriver(t *testing.T) {
	driver := NewDriver("127.0.0.1:3306", "root", "12345678", "test")
	db := driver.Connect()
	res := db.Query("select * from test.test_table")
	fmt.Println(res.Columns[0].Name, res.Columns[1].Name, res.Columns[2].Name)
	for res.Next() {
		fmt.Println(res.GetData(0), res.GetData(1), res.GetData(2))
	}
}
