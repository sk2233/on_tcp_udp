/*
@author: sk
@date: 2024/9/15
*/
package utils

func HandleErr(err error) {
	if err != nil {
		panic(err)
	}
}
