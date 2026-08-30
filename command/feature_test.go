package command

import (
	"testing"

	"github.com/Ptt-Alertor/ptt-alertor/config"
)

func TestDisabledTrackingCommands(t *testing.T) {
	previous := config.Current()
	config.Set(config.Config{Features: config.Features{KeywordTracking: true}})
	t.Cleanup(func() { config.Set(previous) })

	commands := []string{
		"新增作者 gossiping test",
		"新增推文數 gossiping 10",
		"新增噓文數 gossiping 10",
		"新增推文 https://www.ptt.cc/bbs/Gossiping/M.123.A.123.html",
		"推文清單",
		"add -a test gossiping",
		"add -p 10 gossiping",
	}
	for _, input := range commands {
		if got := HandleCommand(input, "test", true); got != DisabledFeatureMessage {
			t.Errorf("HandleCommand(%q) = %q, want %q", input, got, DisabledFeatureMessage)
		}
	}
}

func TestEnabledCommandsOnlyContainsEnabledFeatures(t *testing.T) {
	previous := config.Current()
	config.Set(config.Config{Features: config.Features{KeywordTracking: true}})
	t.Cleanup(func() { config.Set(previous) })

	commands := EnabledCommands()
	if _, ok := commands["關鍵字相關"]; !ok {
		t.Error("keyword commands should be visible")
	}
	for _, category := range []string{"作者相關", "推噓文數相關", "推文相關"} {
		if _, ok := commands[category]; ok {
			t.Errorf("disabled command category %q should be hidden", category)
		}
	}
}

func TestEmptyCommandDoesNotPanic(t *testing.T) {
	if got := HandleCommand("   ", "test", true); got == "" {
		t.Error("empty user command should return guidance")
	}
}
