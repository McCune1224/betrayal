package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
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
	"github.com/mccune1224/betrayal/internal/commands/healthcheck"
	"github.com/mccune1224/betrayal/internal/commands/help"
	"github.com/mccune1224/betrayal/internal/commands/inv"
	"github.com/mccune1224/betrayal/internal/commands/list"
	"github.com/mccune1224/betrayal/internal/commands/roll"
	"github.com/mccune1224/betrayal/internal/commands/search"
	"github.com/mccune1224/betrayal/internal/commands/setup"
	"github.com/mccune1224/betrayal/internal/commands/tarot"
	"github.com/mccune1224/betrayal/internal/commands/view"
	"github.com/mccune1224/betrayal/internal/commands/vote"
	"github.com/mccune1224/betrayal/internal/commands/whisper"
	dbmigrate "github.com/mccune1224/betrayal/internal/db/migrate"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/logger"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/services/datasync"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/mccune1224/betrayal/internal/web"
	"github.com/rs/zerolog"
	"github.com/zekrotja/ken"
	"github.com/zekrotja/ken/state"
)

// config struct to hold env variables and any other config settings
type config struct {
	environment string
	discord     struct {
		clientID       string
		clientSecret   string
		botToken       string
		disableDiscord bool
	}
	database struct {
		dsn string
	}
	web struct {
		port          string
		adminPassword string
		// Railway API
		railwayToken     string
		railwayProjectID string
		railwayServiceID string
		railwayEnvID     string
	}
}

// loadConfig reads the process environment without exposing secret values.
// Only production may select the pooler DSN; local development must provide
// DATABASE_URL explicitly.
func loadConfig(getenv func(string) string) (config, error) {
	env := strings.ToLower(strings.TrimSpace(getenv("ENVIRONMENT")))
	if env == "" {
		// A deployed process must never silently select DATABASE_URL. Local
		// development opts into the direct local database explicitly.
		env = "production"
	}
	var dsnKey string
	switch env {
	case "local":
		dsnKey = "DATABASE_URL"
	case "production":
		dsnKey = "DATABASE_POOLER_URL"
	default:
		return config{}, fmt.Errorf("ENVIRONMENT must be local or production (got %q)", env)
	}
	dsn := getenv(dsnKey)
	if dsn == "" {
		return config{}, fmt.Errorf("%s is required for ENVIRONMENT=%s", dsnKey, env)
	}
	var cfg config
	cfg.environment = env
	cfg.discord.botToken = getenv("DISCORD_BOT_TOKEN")
	cfg.discord.clientID = getenv("DISCORD_CLIENT_ID")
	cfg.discord.clientSecret = getenv("DISCORD_CLIENT_SECRET")
	cfg.discord.disableDiscord = strings.EqualFold(getenv("DISABLE_DISCORD"), "true")
	cfg.database.dsn = dsn
	cfg.web.port = getenv("WEB_PORT")
	if cfg.web.port == "" {
		cfg.web.port = "8080"
	}
	cfg.web.adminPassword = getenv("ADMIN_PASSWORD")
	cfg.web.railwayToken = getenv("RAILWAY_API_TOKEN")
	cfg.web.railwayProjectID = getenv("RAILWAY_BETRAYAL_PROJECT_ID")
	cfg.web.railwayServiceID = getenv("RAILWAY_BETRAYAL_SERVICE_ID")
	cfg.web.railwayEnvID = getenv("RAILWAY_BETRAYAL_ENVIRONMENT_ID")
	return cfg, nil
}

// Global app struct
type app struct {
	dbPool          *pgxpool.Pool
	betrayalManager *ken.Ken
	config          config
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
	cfg, err := loadConfig(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	env := cfg.environment

	// Create the database pool before the logger so the logger can write to it.
	pools, err := pgxpool.New(context.Background(), cfg.database.dsn)
	if err != nil {
		log.Fatalf("Failed to create database connection pool: %v", err)
	}
	defer pools.Close()

	if err := dbmigrate.EnsureUpToDate(cfg.database.dsn); err != nil {
		log.Fatalf("Failed to apply database migrations before startup: %v", err)
	}

	// Initialize the logger exactly once, with database support.
	appLogger, err := logger.Init(logger.Config{
		Environment: env,
		DBPool:      pools,
	})
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	var bot *discordgo.Session

	if !cfg.discord.disableDiscord {
		bot, err = discordgo.New("Bot " + cfg.discord.botToken)
		if err != nil {
			appLogger.Fatal().Err(err).Msg("Error creating Discord session")
		}
		bot.Identify.Intents = gatewayIntents()
	} else {
		appLogger.Info().Msg("DISABLE_DISCORD=true; skipping Discord session startup")
	}

	// Create app instance
	application := &app{
		config: cfg,
		dbPool: pools,
		logger: appLogger,
	}

	// Initialize audit writer
	logger.InitAuditWriter(pools, env)
	defer logger.CloseAuditWriter()

	// Create Ken instance with logger integration when Discord is enabled
	if !cfg.discord.disableDiscord {
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
			},
			OnEventError: func(context string, errMsg error) {
				appLogger.Error().
					Str("event_context", context).
					Err(errMsg).
					Msg("Event error")
			},
		})
		if err != nil {
			appLogger.Fatal().Err(err).Msg("Failed to initialize Ken framework")
		}

		application.betrayalManager = km

		tally := application.RegisterBetrayalCommands(
			new(inv.Inv),
			new(roll.Roll),
			new(action.Action),
			new(view.View),
			new(buy.Buy),
			new(channels.Channel),
			new(help.Help),
			new(vote.Vote),
			new(whisper.Whisper),
			new(whisper.WhisperAdmin),
			new(setup.Setup),
			new(echo.Echo),
			new(list.List),
			new(search.Search),
			new(healthcheck.Healthcheck),
			new(cycle.Cycle),
			new(tarot.Tarot),
		)

		application.betrayalManager.Session().AddHandler(application.logHandler)
		if err := application.betrayalManager.RegisterMiddlewares(logger.NewCommandAuditMiddleware(logger.GetAuditWriter())); err != nil {
			appLogger.Fatal().Err(err).Msg("Failed to register audit middleware")
		}
		application.betrayalManager.Session().AddHandler(paginationHandler)
		defer application.betrayalManager.Unregister()

		if err = bot.Open(); err != nil {
			appLogger.Fatal().Err(err).Msg("Error opening Discord connection")
		}
		defer bot.Close()

		appLogger.Info().
			Str("bot_name", bot.State.User.Username).
			Int("command_count", tally).
			Msg("Bot initialized and running")
	} else {
		appLogger.Info().Msg("Discord functionality disabled; running web server only")
		if strings.Contains(cfg.database.dsn, "roundhouse.proxy.rlwy.net") {
			// The web panel now has state-changing routes (/cycle, player edit,
			// catalog CRUD). In web-only mode against the prod pooler those
			// mutations hit the LIVE game — warn loudly at startup.
			appLogger.Warn().Msg("WEB-ONLY MODE CONNECTED TO PRODUCTION DATABASE (DATABASE_POOLER_URL): changes made in the admin panel affect the live game")
		}
	}

	// Start log retention worker (90 day retention with archival)
	logger.StartRetentionWorker(pools, appLogger, logger.RetentionConfig{
		RetentionDays: 90,
		ArchiveDir:    "./logs_archive",
	})

	// Start web admin server (if password is configured)
	var webServer *web.Server
	if cfg.web.adminPassword != "" {
		var err error
		webServer, err = web.New(pools, bot, appLogger, web.Config{
			Port:             cfg.web.port,
			AdminPassword:    cfg.web.adminPassword,
			DatabaseURL:      cfg.database.dsn,
			Environment:      env,
			SyncEnvURLs:      datasync.EnvURLsFromEnv(),
			RailwayToken:     cfg.web.railwayToken,
			RailwayProjectID: cfg.web.railwayProjectID,
			RailwayServiceID: cfg.web.railwayServiceID,
			RailwayEnvID:     cfg.web.railwayEnvID,
		})
		if err != nil {
			// The Discord bot can still run when the optional web panel has
			// incomplete configuration; do not dereference a nil server after
			// reporting the configuration error.
			appLogger.Error().Err(err).Msg("Failed to initialize web server; web admin disabled")
		} else {
			go func() {
				if err := webServer.Start(); err != nil {
					appLogger.Error().Err(err).Msg("Web server error")
				}
			}()
			appLogger.Info().Str("port", cfg.web.port).Msg("Web admin server started")
		}
	} else {
		appLogger.Warn().Msg("ADMIN_PASSWORD not set, web admin server disabled")
	}

	// Wait for shutdown signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	appLogger.Info().Msg("Shutdown signal received, closing connections")

	// Gracefully shutdown web server
	if webServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := webServer.Shutdown(ctx); err != nil {
			appLogger.Error().Err(err).Msg("Error shutting down web server")
		}
	}

	if application.betrayalManager != nil {
		if err := application.betrayalManager.Session().Close(); err != nil {
			appLogger.Error().Err(err).Msg("Error closing Discord connection")
		}
	}
}

// gatewayIntents returns the gateway intents the bot subscribes with.
//
// IntentsAllWithoutPrivileged keeps the Discord state cache (guilds, channels,
// members, emojis, ...) populated without requiring privileged intents
// (GuildMembers, GuildPresences, MessageContent) to be enabled in the Discord
// developer portal. This replaces the previous assignment of the
// PermissionAdministrator permission constant (value 8 == emoji intent only),
// which starved the state cache.
func gatewayIntents() discordgo.Intent {
	return discordgo.IntentsAllWithoutPrivileged
}

// logHandler logs all slash command invocations to the configured command-log
// channel (see /channel log). If no channel is configured, command logging is
// skipped.
func (a *app) logHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	if i.Member == nil || i.Member.User == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	channelID, err := models.New(a.dbPool).GetCommandLogChannel(ctx)
	if err != nil {
		// Not configured (or DB hiccup) — skip rather than fail the command.
		return
	}

	options := i.ApplicationCommandData().Options
	msg := processOptions(s, options)

	logOutput := fmt.Sprintf("%s - /%s %s - %s", i.Member.User.Username, i.ApplicationCommandData().Name, msg, util.GetEstTimeStamp())

	// Log to Discord channel
	if _, err := s.ChannelMessageSend(channelID, discord.Code(logOutput)); err != nil {
		log.Printf("[CMD] Log send failed: %v", err)
	}
}

// Primary helper for logHandler to process options that a user inputted for a slash command to get invoked (including value arguments)
func processOptions(s *discordgo.Session, options []*discordgo.ApplicationCommandInteractionDataOption) string {
	var msg strings.Builder
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))

	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	for _, opt := range options {
		if o, ok := optionMap[opt.Name]; ok {
			msg.WriteString(formatOption(s, o))
		}
	}

	return msg.String()
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

// paginationHandler handles pagination button interactions for search results
func paginationHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	customID := i.MessageComponentData().CustomID

	// Check if this is a pagination button (format: search_type_userid_timestamp:action or list_type_userid:action)
	if (!strings.Contains(customID, "search_") && !strings.Contains(customID, "list_")) || !strings.Contains(customID, ":") {
		return
	}

	parts := strings.Split(customID, ":")
	if len(parts) != 2 {
		return
	}

	paginationID := parts[0]
	action := parts[1]

	// Get the pagination state
	paginationData := discord.GetPaginationState(paginationID)
	if paginationData == nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "This pagination session has expired. Please run the search command again.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Handle button actions
	switch action {
	case "prev":
		if paginationData.CurrentPage > 0 {
			paginationData.CurrentPage--
			discord.UpdatePaginationState(paginationID, paginationData)
		}

	case "next":
		totalPages := (len(paginationData.Items) + paginationData.PageSize - 1) / paginationData.PageSize
		if paginationData.CurrentPage < totalPages-1 {
			paginationData.CurrentPage++
			discord.UpdatePaginationState(paginationID, paginationData)
		}

	case "done":
		discord.DeletePaginationState(paginationID)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})
		// Delete the message
		s.InteractionResponseDelete(i.Interaction)
		return
	}

	// Create updated embed and components
	embed := discord.CreatePaginatedEmbed(paginationData)
	components := discord.GetPaginationComponents(paginationID, paginationData)

	// Respond with updated message
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}
