package main

import "log"

func main() {
	if err := NewLionWiki(NewSettings()).Run(); err != nil {
		log.Fatal(err)
	}
}
