package workflow

import (
	"testing"
)

func TestParseInfo(t *testing.T) {
	info := DefaultWorkflow().Info()
	if info.BundleID != "net.deanishe.awgo" {
		t.Fatalf("Incorrect bundle ID: %v", info.BundleID)
	}

	if info.Author != "Dean Jackson" {
		t.Fatalf("Incorrect author: %v", info.Author)
	}

	if info.Description != "awgo sample info.plist" {
		t.Fatalf("Incorrect description: %v", info.Description)
	}

	if info.Name != "awgo" {
		t.Fatalf("Incorrect name: %v", info.Name)
	}

	if info.Website != "" {
		t.Fatalf("Incorrect website: %v", info.Website)
	}
}
