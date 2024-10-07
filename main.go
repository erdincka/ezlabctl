/*
Copyright © 2024 Erdinc Kaya <erdincka@msn.com>
*/
package main

import (
	"ezlabctl/cmd"
	"log"
)

func main() {
	// log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetFlags(log.Ltime | log.Lshortfile)
	cmd.Execute()
}
