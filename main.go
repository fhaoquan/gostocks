package main

import (
	"fmt"
	"github.com/myself659/gostocks/collect"
	"time"
)

func main() {
	fmt.Println("hello stock")
	collect.Run()
	<-time.After(time.Second)
}
