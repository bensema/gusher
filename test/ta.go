package main

import (
	"fmt"
	"time"
)

func main() {
	for i := 0; i < 10; i++ {
		t := time.Now().UnixNano()
		fmt.Println(t / 100 % 100000000000)
		fmt.Println(time.Now())
	}

}
