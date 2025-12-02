package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/mccune1224/betrayal/internal/commands/action"
	"github.com/mccune1224/betrayal/internal/commands/buy"
	"github.com/mccune1224/betrayal/internal/commands/channels"
	"github.com/mccune1224/betrayal/internal/commands/cycle"
	"github.com/mccune1224/betrayal/internal/commands/echo"
	"github.com/mccune1224/betrayal/internal/commands/help"
	"github.com/mccune1224/betrayal/internal/commands/inv"
	"github.com/mccune1224/betrayal/internal/commands/list"
	"github.com/mccune1224/betrayal/internal/commands/roll"
	"github.com/mccune1224/betrayal/internal/commands/setup"
	"github.com/mccune1224/betrayal/internal/commands/view"
	"github.com/mccune1224/betrayal/internal/commands/vote"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/logger"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/rs/zerolog"
	"github.com/zekrotja/ken"
	"github.com/zekrotja/ken/state"
)

// config struct to hold env variables and any other config settings
type config struct {
	discord struct {
		clientID     string
		clientSecret string
		botToken     string
	}
	database struct {
		dsn string
	}
}

// Global app struct
type app struct {
	dbPool          *pgxpool.Pool
	betrayalManager *ken.Ken
	conifg          config
	logger          zerolog.Logger
}

// Wrapper for Ken.Command that needs DB access
// (AKA basically every command)
type BetrayalCommand interface {
	ken.Command
	Initialize(*pgxpool.Pool)
}

// Wrapper for ken.RegisterBetrayalCommands for inserting DB access
func (a *app) RegisterBetrayalCommands(commands ...BetrayalCommand) int {
	tally := 0
	for _, command := range commands {
		// command.Initialize(a.dbPool, &a.scheduler)
		command.Initialize(a.dbPool)
		err := a.betrayalManager.RegisterCommands(command)
		if err != nil {
			log.Fatal(err)
		}
		tally += 1

	}
	return tally
}

func main() {
	// Initialize logger
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "local"
	}
	appLogger, err := logger.Init(logger.Config{Environment: env})
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	var cfg config
	cfg.discord.botToken = os.Getenv("DISCORD_BOT_TOKEN")
	cfg.discord.clientID = os.Getenv("DISCORD_CLIENT_ID")
	cfg.discord.clientSecret = os.Getenv("DISCORD_CLIENT_SECRET")
	cfg.database.dsn = os.Getenv("DATABASE_POOLER_URL")

	// Spin up Bot and give it admin permissions
	bot, err := discordgo.New("Bot " + cfg.discord.botToken)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Error creating Discord session")
	}
	bot.Identify.Intents = discordgo.PermissionAdministrator

	// Create database pool
	pools, err := pgxpool.New(context.Background(), cfg.database.dsn)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to create database connection pool")
	}
	defer pools.Close()

	// Create app instance
	application := &app{
		conifg: cfg,
		dbPool: pools,
		logger: appLogger,
	}

	// Initialize audit writer
	logger.InitAuditWriter(pools, env)
	defer logger.CloseAuditWriter()

	// Create Ken instance with logger integration
	km, err := ken.New(bot, ken.Options{
		State: state.NewInternal(),
		EmbedColors: ken.EmbedColors{
			Default: discord.ColorThemeOrange,
			Error:   discord.ColorThemeRuby,
		},
		DisableCommandInfoCache: true,
		OnSystemError: func(ctx string, errMsg error, args ...any) {
			appLogger.Error().
				Str("context", ctx).
				Err(errMsg).
				Any("args", args).
				Msg("System error")
		},
		OnCommandError: func(errMsg error, ctx *ken.Ctx) {
			logger.InjectKenContext(ctx)
			cmdLogger := logger.FromKenContext(ctx)

			cmdArg := processOptions(bot, ctx.GetEvent().ApplicationCommandData().Options)
			cmdLogger.Error().
				Err(errMsg).
				Str("options", cmdArg).
				Msg("Command execution failed")

			// Log to audit for failed commands
			auditWriter := logger.GetAuditWriter()
			if auditWriter != nil && ctx.GetEvent().Member != nil && ctx.GetEvent().Member.User != nil {
				audit := logger.CreateAuditFromContext(ctx, bot, time.Now())
				audit.Status = "error"
				errorMsg := errMsg.Error()
				audit.ErrorMessage = &errorMsg
				audit.ExecutionTimeMs = 0 // Unknown for error case
				auditWriter.LogCommand(audit)
			}
		},
		OnEventError: func(context string, errMsg error) {
			appLogger.Error().
				Str("event_context", context).
				Err(errMsg).
				Msg("Event error")
		},
	})
	application.betrayalManager = km
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to initialize Ken framework")
	}

	// Call unregister twice to remove any lingering commands from previous runs
	application.betrayalManager.Unregister()

	tally := application.RegisterBetrayalCommands(
		new(inv.Inv),
		new(roll.Roll),
		new(action.Action),
		new(view.View),
		new(buy.Buy),
		new(channels.Channel),
		new(help.Help),
		new(vote.Vote),
		new(setup.Setup),
		new(echo.Echo),
		new(list.List),
		new(cycle.Cycle),
	)

	application.betrayalManager.Session().AddHandler(logHandler)
	application.betrayalManager.Session().AddHandler(auditHandler)
	defer application.betrayalManager.Unregister()

	err = bot.Open()
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Error opening Discord connection")
	}
	defer bot.Close()

	appLogger.Info().
		Str("bot_name", bot.State.User.Username).
		Int("command_count", tally).
		Msg("Bot initialized and running")

	// Start log retention worker (90 day retention with archival)
	logger.StartRetentionWorker(pools, appLogger, logger.RetentionConfig{
		RetentionDays: 90,
		ArchiveDir:    "./logs_archive",
	})

	// Wait for shutdown signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	appLogger.Info().Msg("Shutdown signal received, closing connections")
	if err := application.betrayalManager.Session().Close(); err != nil {
		appLogger.Error().Err(err).Msg("Error closing Discord connection")
	}
}

// TODO: Make Log Channel configurable with a slash command maybe?

// Handles logging all slash commands to a dedicated channel
func logHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	testLoggerID := "1108318770138714163"
	options := i.ApplicationCommandData().Options
	msg := processOptions(s, options)

	logOutput := fmt.Sprintf("%s - /%s %s - %s", i.Member.User.Username, i.ApplicationCommandData().Name, msg, util.GetEstTimeStamp())

	// Log to Discord channel
	_, err := s.ChannelMessageSend(testLoggerID, discord.Code(logOutput))
	if err != nil {
		log.Printf("[CMD] Log send failed: %v", err)
	}
}

// Primary helper for logHandler to process options that a user inputted for a slash command to get invoked (including value arguments)
func processOptions(s *discordgo.Session, options []*discordgo.ApplicationCommandInteractionDataOption) string {
	var msg string
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))

	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	for _, opt := range options {
		if o, ok := optionMap[opt.Name]; ok {
			msg += formatOption(s, o)
		}
	}

	return msg
}

// Helper to handle parsing argument options to a string format (really should just be apart of processOptions but...Too Bad!)
// Also this function is a mess, but it works even if I'm using recursion :)
func formatOption(s *discordgo.Session, o *discordgo.ApplicationCommandInteractionDataOption) string {
	switch o.Type {
	default:
		return ""
	case discordgo.ApplicationCommandOptionString:
		return fmt.Sprintf("%s:%s, ", o.Name, o.StringValue())
	case discordgo.ApplicationCommandOptionInteger:
		return fmt.Sprintf("%s:%d, ", o.Name, o.IntValue())
	case discordgo.ApplicationCommandOptionBoolean:
		return fmt.Sprintf("%s:%t, ", o.Name, o.BoolValue())
	case discordgo.ApplicationCommandOptionUser:
		return fmt.Sprintf("%s:%s, ", o.Name, o.UserValue(s).Username)
	case discordgo.ApplicationCommandOptionChannel:
		return fmt.Sprintf("%s:%s, ", o.Name, o.ChannelValue(s).Name)
	case discordgo.ApplicationCommandOptionSubCommand:
		return fmt.Sprintf("%s %s", o.Name, processOptions(s, o.Options))
	case discordgo.ApplicationCommandOptionSubCommandGroup:
		return fmt.Sprintf("%s %s", o.Name, processOptions(s, o.Options))

		// I don't think there's ever going to be a case where I'm using these...
	case discordgo.ApplicationCommandOptionRole:
		return ""
	case discordgo.ApplicationCommandOptionMentionable:
		return ""
	}
	// new info
}

// auditHandler logs all successful slash commands to the database for audit trail
func auditHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	// Only log successful command executions (those without errors are handled after completion)
	// This handler catches the interaction before processing
	auditWriter := logger.GetAuditWriter()
	if auditWriter == nil || i.Member == nil || i.Member.User == nil {
		return
	}

	cmdData := i.ApplicationCommandData()
	arguments := logger.ExtractCommandArguments(s, cmdData.Options)

	userRoles := i.Member.Roles
	isAdmin := false
	for _, roleID := range userRoles {
		for _, adminRole := range discord.AdminRoles {
			if roleID == adminRole {
				isAdmin = true
				break
			}
		}
	}

	audit := logger.CommandAudit{
		CorrelationID:    logger.GenerateCorrelationID().String(),
		CommandName:      cmdData.Name,
		UserID:           i.Member.User.ID,
		Username:         i.Member.User.Username,
		UserRoles:        userRoles,
		GuildID:          i.GuildID,
		ChannelID:        i.ChannelID,
		IsAdmin:          isAdmin,
		CommandArguments: arguments,
		Status:           "success",
		ExecutionTimeMs:  0, // Will be updated if we can track it
	}

	auditWriter.LogCommand(audit)
}
