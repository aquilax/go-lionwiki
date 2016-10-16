package main

import "log"

func main() {
	if err := NewLionWiki().Run(NewSettings()); err != nil {
		log.Fatal(err)
	}
}
