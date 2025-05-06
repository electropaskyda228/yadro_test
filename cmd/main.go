package main

import (
	"fmt"
	"os"

	. "biatlon/internals"
)

func main() {
	fmt.Println("start script")
	race, err := GetConfiguration("config.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ProccessRace(race, "events", "outlog.txt")
	fmt.Println("end script")
}
