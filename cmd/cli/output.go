package cli

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/pretty"
)

func outputJSON(data any) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if RawFlag {
		prettyJSON := pretty.Pretty(jsonBytes)
		fmt.Print(string(prettyJSON))
		if len(prettyJSON) > 0 && prettyJSON[len(prettyJSON)-1] != '\n' {
			fmt.Println()
		}
		return nil
	}

	prettyJSON := pretty.Pretty(jsonBytes)
	coloredJSON := pretty.Color(prettyJSON, nil)
	fmt.Print(string(coloredJSON))
	if len(coloredJSON) > 0 && coloredJSON[len(coloredJSON)-1] != '\n' {
		fmt.Println()
	}
	return nil
}
