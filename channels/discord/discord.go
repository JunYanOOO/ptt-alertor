package discord

import (
	"errors"
	"os"
	"strings"
	"sync"

	log "github.com/Ptt-Alertor/logrus"
	"github.com/bwmarrin/discordgo"

	"github.com/Ptt-Alertor/ptt-alertor/command"
	"github.com/Ptt-Alertor/ptt-alertor/myutil"
)

// SplitTextByLineBreak may include the boundary rune in a chunk, so keeping
// one character in reserve guarantees Discord's 2,000-character limit.
const maxCharacters = 1999

var (
	token   = os.Getenv("DISCORD_TOKEN")
	guildID = os.Getenv("DISCORD_GUILD_ID")

	mu  sync.RWMutex
	bot *discordgo.Session
)

// Start connects the Discord bot and registers its slash commands. When
// DISCORD_GUILD_ID is set, commands are registered only in that server and are
// available immediately. Without it, Discord registers them globally.
func Start() error {
	if token == "" {
		log.Info("Discord Bot Disabled: DISCORD_TOKEN is not set")
		return nil
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return err
	}
	session.Identify.Intents = discordgo.IntentsGuilds
	session.AddHandler(handleInteraction)

	if err := session.Open(); err != nil {
		return err
	}

	mu.Lock()
	bot = session
	mu.Unlock()

	for _, applicationCommand := range applicationCommands() {
		if _, err := session.ApplicationCommandCreate(session.State.User.ID, guildID, applicationCommand); err != nil {
			Close()
			return err
		}
	}

	registrationScope := "globally"
	if guildID != "" {
		registrationScope = "in guild " + guildID
	}
	log.WithField("scope", registrationScope).Info("Discord Bot Started")
	return nil
}

// Enabled reports whether the Discord gateway connection is active.
func Enabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return bot != nil
}

// Close closes the Discord gateway connection.
func Close() {
	mu.Lock()
	session := bot
	bot = nil
	mu.Unlock()
	if session != nil {
		if err := session.Close(); err != nil {
			log.WithError(err).Error("Discord Bot Close Failed")
		}
	}
}

func applicationCommands() []*discordgo.ApplicationCommand {
	manageChannels := int64(discordgo.PermissionManageChannels)
	dmPermission := false
	boardOption := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "看板",
		Description: "PTT 看板名稱，例如 gossiping",
		Required:    true,
	}
	keywordOption := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "關鍵字",
		Description: "文章標題要包含的關鍵字",
		Required:    true,
	}

	return []*discordgo.ApplicationCommand{
		{
			Name:                     "新增",
			Description:              "新增 PTT 看板關鍵字追蹤",
			DefaultMemberPermissions: &manageChannels,
			DMPermission:             &dmPermission,
			Options:                  []*discordgo.ApplicationCommandOption{boardOption, keywordOption},
		},
		{
			Name:                     "清單",
			Description:              "查看此頻道的 PTT 追蹤清單",
			DefaultMemberPermissions: &manageChannels,
			DMPermission:             &dmPermission,
		},
		{
			Name:                     "刪除",
			Description:              "刪除此頻道的 PTT 看板關鍵字追蹤",
			DefaultMemberPermissions: &manageChannels,
			DMPermission:             &dmPermission,
			Options:                  []*discordgo.ApplicationCommandOption{boardOption, keywordOption},
		},
	}
}

func handleInteraction(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if interaction.Type != discordgo.InteractionApplicationCommand {
		return
	}

	if interaction.GuildID == "" {
		respond(session, interaction, "請在 Discord 伺服器的文字頻道使用此指令。")
		return
	}

	if err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
	}); err != nil {
		log.WithError(err).Error("Discord Interaction Defer Failed")
		return
	}

	response, err := executeCommand(interaction)
	if err != nil {
		response = "操作失敗：" + err.Error()
	}
	response = truncate(response, maxCharacters)
	if _, err := session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{Content: &response}); err != nil {
		log.WithError(err).Error("Discord Interaction Response Failed")
	}
}

func executeCommand(interaction *discordgo.InteractionCreate) (string, error) {
	account, err := command.HandleDiscordFollow(interaction.ChannelID)
	if err != nil {
		return "", err
	}

	data := interaction.ApplicationCommandData()
	switch data.Name {
	case "新增", "刪除":
		if len(data.Options) != 2 {
			return "", errors.New("看板與關鍵字都是必填欄位")
		}
		options := make(map[string]string, len(data.Options))
		for _, option := range data.Options {
			options[option.Name] = strings.TrimSpace(option.StringValue())
		}
		if options["看板"] == "" || options["關鍵字"] == "" {
			return "", errors.New("看板與關鍵字不可空白")
		}
		return command.HandleCommand(data.Name+" "+options["看板"]+" "+options["關鍵字"], account, true), nil
	case "清單":
		return command.HandleCommand("清單", account, true), nil
	default:
		return "", errors.New("不支援的指令")
	}
}

func respond(session *discordgo.Session, interaction *discordgo.InteractionCreate, content string) {
	err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.WithError(err).Error("Discord Interaction Response Failed")
	}
}

// SendTextMessage sends an alert to a subscribed Discord channel.
func SendTextMessage(channelID, content string) {
	mu.RLock()
	session := bot
	mu.RUnlock()
	if session == nil {
		log.Warn("Discord message skipped: integration is disabled")
		return
	}

	for _, message := range myutil.SplitTextByLineBreak(content, maxCharacters) {
		if _, err := session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
			Content: message,
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{},
			},
		}); err != nil {
			log.WithField("channel", channelID).WithError(err).Error("Discord Send Message Failed")
		}
	}
}

func truncate(content string, limit int) string {
	characters := []rune(content)
	if len(characters) <= limit {
		return content
	}
	return string(characters[:limit-1]) + "…"
}
