package main

import (
	"fmt"
	"time"
)

func main() {
	a := time.UnixMicro(1776351478547178)
	b := time.UnixMicro(1776351478901450)

	dur := b.Sub(a)
	fmt.Println(dur.String())
}
