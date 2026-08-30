package discord

import (
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestApplicationCommands(t *testing.T) {
	commands := applicationCommands()
	if len(commands) != 3 {
		t.Fatalf("applicationCommands() returned %d commands, want 3", len(commands))
	}

	wantNames := []string{"新增", "清單", "刪除"}
	for index, applicationCommand := range commands {
		if applicationCommand.Name != wantNames[index] {
			t.Errorf("command %d name = %q, want %q", index, applicationCommand.Name, wantNames[index])
		}
		if applicationCommand.DefaultMemberPermissions == nil || *applicationCommand.DefaultMemberPermissions != discordgo.PermissionManageChannels {
			t.Errorf("command %q must require Manage Channels", applicationCommand.Name)
		}
		if applicationCommand.DMPermission == nil || *applicationCommand.DMPermission {
			t.Errorf("command %q must be disabled in direct messages", applicationCommand.Name)
		}
	}
}

func TestTruncate(t *testing.T) {
	content := strings.Repeat("新", 10)
	if got := truncate(content, 5); got != "新新新新…" {
		t.Errorf("truncate() = %q, want %q", got, "新新新新…")
	}
	if got := truncate("PTT", 5); got != "PTT" {
		t.Errorf("truncate() changed short content to %q", got)
	}
}
