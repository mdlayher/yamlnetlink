package main

import (
	"encoding/json"
	"log"
	"os"
)

func main() {
	c, err := Dial(nil)
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}
	defer c.Close()

	family, err := c.DoGetfamily(DoGetfamilyRequest{FamilyName: "nlctrl"})
	if err != nil {
		log.Fatalf("failed to get nlctrl: %v", err)
	}

	all, err := c.DumpGetfamily()
	if err != nil {
		log.Fatalf("failed to dump families: %v", err)
	}

	families := make([]string, 0, len(all))
	for _, f := range all {
		families = append(families, f.FamilyName)
	}

	_ = json.NewEncoder(os.Stdout).Encode(stdout{
		Family:   *family,
		Families: families,
	})
}

type stdout struct {
	Family   DoGetfamilyReply `json:"family"`
	Families []string         `json:"families"`
}
