package main

func main() {
	if err := NewLionWiki().Run(); err != nil {
		panic(err)
	}
}
