package prompts

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ConfirmDeletion prompts user to confirm a deletion operation
// Returns true if user confirms, false otherwise
func ConfirmDeletion(resourceType, resourceName, cascadeMessage string) bool {
	fmt.Printf("⚠ This will delete %s '%s'", resourceType, resourceName)
	if cascadeMessage != "" {
		fmt.Printf(" and %s", cascadeMessage)
	}
	fmt.Println()
	fmt.Print("Are you sure? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
