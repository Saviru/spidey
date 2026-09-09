package sandbox

import (
	"testing"
)

func TestSandbox(t *testing.T) {
	// Test allowed code
	safeCode := `
import "fmt"
func Render() map[string]interface{} { return nil }
`
	if err := ValidateFrontmatter(safeCode); err != nil {
		t.Fatalf("Safe code should pass, got error: %v", err)
	}

	// Test blocked code
	unsafeCode := `
import "os/exec"
func Render() map[string]interface{} { return nil }
`
	if err := ValidateFrontmatter(unsafeCode); err == nil {
		t.Fatal("Blocked code should return an error, but got nil")
	}

	// Test merging parameters
	renderData := map[string]interface{}{"Role": "Admin"}
	urlParams := map[string]interface{}{"Id": "42"}

	merged := MergeParams(renderData, urlParams).(map[string]interface{})
	if merged["Id"] != "42" || merged["Role"] != "Admin" {
		t.Fatalf("Failed to merge params correctly: %+v", merged)
	}

	t.Log("Sandbox tests passed successfully")
}
