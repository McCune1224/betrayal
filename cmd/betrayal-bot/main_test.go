package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestLoadConfigSelectsLocalDatabaseURL(t *testing.T) {
	cfg, err := loadConfig(func(key string) string {
		return map[string]string{"ENVIRONMENT": "local", "DATABASE_URL": "postgres://local/db?sslmode=disable", "DATABASE_POOLER_URL": "postgres://production/db"}[key]
	})
	if err != nil {
		t.Fatalf("load local config: %v", err)
	}
	if cfg.database.dsn != "postgres://local/db?sslmode=disable" {
		t.Fatalf("local config selected %q; want DATABASE_URL", cfg.database.dsn)
	}
}

func TestLoadConfigSelectsProductionPoolerURL(t *testing.T) {
	cfg, err := loadConfig(func(key string) string {
		return map[string]string{"ENVIRONMENT": "production", "DATABASE_URL": "postgres://local/db", "DATABASE_POOLER_URL": "postgres://production/db"}[key]
	})
	if err != nil {
		t.Fatalf("load production config: %v", err)
	}
	if cfg.database.dsn != "postgres://production/db" {
		t.Fatalf("production config selected %q; want DATABASE_POOLER_URL", cfg.database.dsn)
	}
}

func TestLoadConfigFailsClosedWhenRequiredDatabaseURLIsMissing(t *testing.T) {
	_, err := loadConfig(func(key string) string {
		return map[string]string{"ENVIRONMENT": "local", "DATABASE_POOLER_URL": "postgres://production/db"}[key]
	})
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("expected missing local DATABASE_URL error, got %v", err)
	}
}

func TestLoadConfigFailsClosedWhenProductionPoolerURLIsMissing(t *testing.T) {
	_, err := loadConfig(func(key string) string { return map[string]string{"ENVIRONMENT": "production"}[key] })
	if err == nil || !strings.Contains(err.Error(), "DATABASE_POOLER_URL") {
		t.Fatalf("expected missing production DSN error, got %v", err)
	}
}

func TestMakefileUsesExplicitProductionMigrationConfirmation(t *testing.T) {
	makefile, err := os.ReadFile(filepath.Join("..", "..", "Makefile"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(makefile)
	for _, want := range []string{"migrate-local-up:", "migrate-production-up:", "CONFIRM_PRODUCTION_MIGRATION=YES"} {
		if !strings.Contains(text, want) {
			t.Errorf("Makefile missing %q", want)
		}
	}
	if strings.Contains(text, "migrate-up:\n	migrate -database $(call env-value,DATABASE_POOLER_URL)") {
		t.Fatal("ambiguous migrate-up target still points directly at production")
	}
}

func TestGatewayIntents(t *testing.T) {
	intents := gatewayIntents()

	// Regression: the intents field used to be assigned the
	// PermissionAdministrator permission constant (1<<3 == 8), which happens to
	// equal the emoji intent bit — a permission, not a real intent set.
	if int(intents) == int(discordgo.PermissionAdministrator) {
		t.Fatal("gateway intents must not be the PermissionAdministrator permission constant")
	}

	// Intents needed to keep the Discord state cache populated and to observe
	// message/reaction activity in the bot's guilds.
	required := []discordgo.Intent{
		discordgo.IntentsGuilds,
		discordgo.IntentsGuildMessages,
		discordgo.IntentsGuildMessageReactions,
		discordgo.IntentsGuildEmojis,
		discordgo.IntentsGuildVoiceStates,
	}
	for _, want := range required {
		if intents&want != want {
			t.Errorf("gateway intents missing bit %d", want)
		}
	}

	// Privileged intents (GuildMembers, GuildPresences, MessageContent) require
	// explicit enablement in the Discord developer portal; a bot that connects
	// without them enabled gets disconnected with a 4014. IntentsAllWithoutPrivileged
	// deliberately excludes them.
	privileged := discordgo.IntentsGuildMembers | discordgo.IntentsGuildPresences | discordgo.IntentsMessageContent
	if intents&privileged != 0 {
		t.Errorf("gateway intents unexpectedly include privileged bits %d", intents&privileged)
	}
}
