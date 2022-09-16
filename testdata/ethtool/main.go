package main

import (
	"fmt"
	"log"

	"github.com/mdlayher/netlink"
)

func main() {
	c, err := Dial(&netlink.Config{Strict: true})
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}
	defer c.Close()

	// TODO(mdlayher): this doesn't quite work yet.
	fmt.Println("{}")
}
