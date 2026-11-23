package app

import (
	"context"
	"errors"
	"sync"
)

const ChristmasIntroduction = `The air got colder, and colder, and colder at the Milkyway Galaxy.
A very faint bells sound could be heard at the distance...
...Ho...Ho...Ho...what a merry night...
Looking at the distance we see lights approaching!
What is that!? It's a Christmas party starting at the Milkyway Galaxy!`

const GuildSetupPrompt = "Welcome to Milky Xmas! Moderators, use /setup channel to choose where commands work and where the Grinch can appear. Use /setup view to see your choices."

var ErrManageGuild = errors.New("Manage Server permission is required")

type ChannelSettings struct {
	Commands   bool
	Boss       bool
	Introduced bool
}

type GuildSettings struct {
	Welcomed bool
	Channels map[string]ChannelSettings
}

type GuildSettingsStore interface {
	LoadGuildSettings(context.Context, string) (GuildSettings, error)
	SaveGuildSettings(context.Context, string, GuildSettings) error
}

type SetupAnnouncement func(context.Context, string, string) error

type GuildSetup struct {
	store GuildSettingsStore
	locks sync.Map
}

func NewGuildSetup(store GuildSettingsStore) *GuildSetup {
	return &GuildSetup{store: store}
}

func (setup *GuildSetup) lock(guild string) func() {
	value, _ := setup.locks.LoadOrStore(guild, &sync.Mutex{})
	mu := value.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

func (setup *GuildSetup) Settings(ctx context.Context, guild string) (GuildSettings, error) {
	return setup.store.LoadGuildSettings(ctx, guild)
}

func (setup *GuildSetup) Welcome(ctx context.Context, guild, channel string, post SetupAnnouncement) error {
	unlock := setup.lock(guild)
	defer unlock()
	settings, err := setup.Settings(ctx, guild)
	if err != nil || settings.Welcomed || len(settings.Channels) != 0 {
		return err
	}
	if err := post(ctx, channel, GuildSetupPrompt); err != nil {
		return err
	}
	settings.Welcomed = true
	return setup.store.SaveGuildSettings(ctx, guild, settings)
}

// Configure reports whether the channel choices were saved, even if posting failed.
func (setup *GuildSetup) Configure(ctx context.Context, guild, channel string, canManage, commands, boss bool, post SetupAnnouncement) (bool, error) {
	if !canManage {
		return false, ErrManageGuild
	}

	unlock := setup.lock(guild)
	defer unlock()
	settings, err := setup.Settings(ctx, guild)
	if err != nil {
		return false, err
	}
	if settings.Channels == nil {
		settings.Channels = make(map[string]ChannelSettings)
	}
	selection := settings.Channels[channel]
	selection.Commands, selection.Boss = commands, boss
	if !commands && !boss && !selection.Introduced {
		delete(settings.Channels, channel)
	} else {
		settings.Channels[channel] = selection
	}
	if err := setup.store.SaveGuildSettings(ctx, guild, settings); err != nil {
		return false, err
	}
	if !boss || selection.Introduced {
		return true, nil
	}
	if err := post(ctx, channel, ChristmasIntroduction); err != nil {
		return true, err
	}
	selection.Introduced = true
	settings.Channels[channel] = selection
	return true, setup.store.SaveGuildSettings(ctx, guild, settings)
}
