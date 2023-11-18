package commands

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/scheduler"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
)

type Test struct {
	models    data.Models
	scheduler *scheduler.BetrayalScheduler
}

var _ ken.SlashCommand = (*Test)(nil)

// Initialize implements BetrayalCommand.
func (t *Test) Initialize(m data.Models, s *scheduler.BetrayalScheduler) {
	t.models = m
	t.scheduler = s
}

// Description implements ken.SlashCommand.
func (*Test) Description() string {
	return "Dev Sandbox for commands"
}

// Name implements ken.SlashCommand.
func (*Test) Name() string {
	return "t"
}

// Options implements ken.SlashCommand.
func (*Test) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "timer",
			Description: "test cron job timer",
			Options: []*discordgo.ApplicationCommandOption{
				discord.StringCommandArg("duration", "duration of the timer", true),
				discord.StringCommandArg("task", "Task to do", true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "queue",
			Description: "see queue for a user",
			Options: []*discordgo.ApplicationCommandOption{
				discord.UserCommandArg(true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "ac",
			Description: "test autocorrect",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "status",
					Description: "status fzf and levenstein distance",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("status", "status to autocorrect", true),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "role",
					Description: "role fzf and levenstein distance",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("role", "role to autocorrect", true),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "ability",
					Description: "ability fzf and levenstein distance",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("ability", "ability to autocorrect", true),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "aa",
					Description: "ability fzf and levenstein distance",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("aa", "aa to autocorrect", true),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "perk",
					Description: "perk fzf and levenstein distance",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("perk", "status to autocorrect", true),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "item",
					Description: "item fzf and levenstein distance",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("item", "item to autocorrect", true),
					},
				},
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (t *Test) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "timer", Run: t.testTimer},
		ken.SubCommandHandler{Name: "queue", Run: t.testQueue},
		ken.SubCommandGroup{Name: "ac", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "role", Run: t.testAcRole},
			ken.SubCommandHandler{Name: "status", Run: t.testAcStatus},
			ken.SubCommandHandler{Name: "ability", Run: t.testAcAbility},
			ken.SubCommandHandler{Name: "aa", Run: t.testAcAA},
			ken.SubCommandHandler{Name: "item", Run: t.testAcItem},
		}},
	)
}

// Version implements ken.SlashCommand.
func (*Test) Version() string {
	return "1.0.0"
}

func (t *Test) testTimer(ctx ken.SubCommandContext) (err error) {
	dur := ctx.Options().GetByName("duration").StringValue()
	timeDur, err := time.ParseDuration(dur)
	if err != nil {
		return discord.ErrorMessage(ctx, "Failed to Parse Duration Argument", `A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`)
	}
	taskArg := ctx.Options().GetByName("task").StringValue()

	// WARNING: Is this the correct way to do this? I don't know
	// Do care? No
	// Anon functions are fun
	job := func() { sendMessageTask(ctx.GetSession(), ctx.GetEvent(), taskArg) }
	err = t.scheduler.UpsertJob(time.Now().String(), timeDur, job)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to schedule timer")
	}
	return discord.SuccessfulMessage(ctx, "Timer Scheduled", fmt.Sprintf("Timer scheduled for %s", timeDur.String()))
}

// Helper function to test sending a cron job message with the correct context
func sendMessageTask(s *discordgo.Session, e *discordgo.InteractionCreate, arg string) {
	ch := e.ChannelID
	msg := fmt.Sprintf("Hello from cron job: %s", arg)
	_, err := s.ChannelMessageSend(ch, msg)
	if err != nil {
		log.Println(err)
	}
}

// Helper struct to handle ac tests
type acResult struct {
	Arg   string
	Best  string
	Time  time.Duration
	Count int
}

func (t *Test) testAcStatus(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	statusArg := ctx.Options().GetByName("status").StringValue()
	// fuzzy matching
	start := time.Now()
	best, err := t.models.Statuses.GetFuzzy(statusArg)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Failed to get statuses", err.Error())
	}
	stop := time.Now()

	total := stop.Sub(start)
	totalTime := total.Nanoseconds()
	msg := fmt.Sprintf("%s => %s %dns", statusArg, best.Name, totalTime)
	return ctx.RespondMessage(discord.Code(msg))
}

func (t *Test) testAcRole(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	roleArg := ctx.Options().GetByName("role").StringValue()
	roles, err := t.models.Roles.GetAll()
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Failed to get statuses", "DB error idk lol")
	}

	names := make([]string, len(roles))
	for i := range roles {
		names[i] = roles[i].Name
	}

	// fuzzy matching
	start := time.Now()
	best, _ := util.FuzzyFind(roleArg, names)
	stop := time.Now()

	total := stop.Sub(start)
	totalTime := total.Nanoseconds()
	msg := fmt.Sprintf("%s => %s Searched %d  %dns", roleArg, best, len(names), totalTime)
	return ctx.RespondMessage(discord.Code(msg))
}

func (t *Test) testAcAbility(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	abilityArg := ctx.Options().GetByName("ability").StringValue()
	// fuzzy matching
	start := time.Now()
	best, err := t.models.Abilities.GetByFuzzy(abilityArg)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Failed to get statuses", err.Error())
	}
	stop := time.Now()

	total := stop.Sub(start)
	totalTime := total.Nanoseconds()
	msg := fmt.Sprintf("%s => %s %dns", abilityArg, best.Name, totalTime)
	return ctx.RespondMessage(discord.Code(msg))
}

func (t *Test) testAcAA(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	aaArg := ctx.Options().GetByName("aa").StringValue()
	aas, err := t.models.Abilities.GetAllAnyAbilities()
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Failed to get statuses", "DB error idk lol")
	}

	names := make([]string, len(aas))
	for i := range aas {
		names[i] = aas[i].Name
	}

	// fuzzy matching
	start := time.Now()
	best, _ := util.FuzzyFind(aaArg, names)
	stop := time.Now()

	total := stop.Sub(start)
	totalTime := total.Nanoseconds()
	msg := fmt.Sprintf("%s => %s Searched %d  %dns", aaArg, best, len(names), totalTime)
	return ctx.RespondMessage(discord.Code(msg))
}

func (t *Test) testAcItem(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	itemArg := ctx.Options().GetByName("item").StringValue()
	items, err := t.models.Items.GetAll()
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Failed to get statuses", "DB error idk lol")
	}

	names := make([]string, len(items))
	for i := range items {
		names[i] = items[i].Name
	}

	// fuzzy matching
	start := time.Now()
	best, _ := util.FuzzyFind(itemArg, names)
	stop := time.Now()

	total := stop.Sub(start)
	totalTime := total.Nanoseconds()
	msg := fmt.Sprintf("%s => %s Searched %d  %dns", itemArg, best, len(names), totalTime)
	return ctx.RespondMessage(discord.Code(msg))
}

func (t *Test) testQueue(ctx ken.SubCommandContext) (err error) {
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.ErrorMessage(ctx, "You are not an admin", "You must be an admin to use this command")
	}
	userArg := ctx.Options().GetByName("user").UserValue(ctx)

	userInv, err := t.models.Inventories.GetByDiscordID(userArg.ID)
	if err != nil {
		return discord.ErrorMessage(ctx, "Unable to find inventory", "Unable to find inventory for user")
	}
	userJobs, err := t.models.InventoryCronJobs.GetByInventoryID(userInv.DiscordID)
	if userJobs != nil {
		return ctx.RespondMessage("User has no jobs")
	}

	fields := []*discordgo.MessageEmbedField{}
	for _, job := range userJobs {
		// convert job.InvokeTime to time
		invokes := time.Unix(job.InvokeTime, 0)
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   job.JobID,
			Value:  fmt.Sprintf("Scheduled for %s", invokes.String()),
			Inline: true,
		})
	}
	msg := discordgo.MessageEmbed{
		Title:       "Jobs for " + userArg.Username,
		Description: fmt.Sprintf("total jobs: %d", len(userJobs)),
		Fields:      fields,
	}

	return ctx.RespondEmbed(&msg)
}
