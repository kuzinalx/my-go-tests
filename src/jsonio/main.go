package main

import (
	"encoding/json"
	"fmt"
)

type Small struct {
	A int
}

type Big struct {
	A int
	B int
}

func main() {
	small := Small{A: 1}
	buf, _ := json.Marshal(&small)
	var big Big
	big.B = 2
	err := json.Unmarshal(buf, &big)
	if err != nil {
		fmt.Print(err)
	} else {
		fmt.Print(big)
	}
	go func() {
		for {
			fmt.Print("ggg")
		}
	}()
}
