package main

import "log"

func main() {
	if err := NewLionWiki().Run(); err != nil {
		log.Fatal(err)
	}
}
