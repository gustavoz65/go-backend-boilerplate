package utils

import (
	"encoding/json"
	"fmt"
)

func PrintJSON(c interface{}) {
	json, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}
	fmt.Println("JSON:", string(json))
}
