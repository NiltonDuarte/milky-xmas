package discord

import (
	"context"
	"sync"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/bwmarrin/discordgo"
)

type grinchChannel struct{ guild, channel string }

type grinchWorker struct {
	bot  *grinchBot
	stop context.CancelFunc
}

type grinchChannels struct {
	mu      sync.Mutex
	wg      sync.WaitGroup
	ctx     context.Context
	setup   *app.GuildSetup
	build   func(string, string) *app.GrinchEncounter
	workers map[grinchChannel]grinchWorker
}

func (channels *grinchChannels) refresh(ctx context.Context, session *discordgo.Session, guild string) error {
	channels.mu.Lock()
	defer channels.mu.Unlock()
	settings, err := channels.setup.Settings(ctx, guild)
	if err != nil {
		return err
	}
	for id, worker := range channels.workers {
		if id.guild == guild && !settings.Channels[id.channel].Boss {
			worker.stop()
			delete(channels.workers, id)
		}
	}
	if channels.ctx.Err() != nil {
		return channels.ctx.Err()
	}
	for channel, settings := range settings.Channels {
		id := grinchChannel{guild, channel}
		if !settings.Boss || channels.workers[id].bot != nil {
			continue
		}
		workerCtx, stop := context.WithCancel(channels.ctx)
		worker := &grinchBot{guild: guild, channel: channel, encounter: channels.build(guild, guild+":"+channel)}
		channels.workers[id] = grinchWorker{bot: worker, stop: stop}
		channels.wg.Go(func() { worker.run(workerCtx, session) })
	}
	return nil
}

func (channels *grinchChannels) get(guild, channel string) *grinchBot {
	if channels == nil {
		return nil
	}
	channels.mu.Lock()
	defer channels.mu.Unlock()
	return channels.workers[grinchChannel{guild, channel}].bot
}

func (channels *grinchChannels) remove(guild string) {
	channels.mu.Lock()
	defer channels.mu.Unlock()
	for id, worker := range channels.workers {
		if id.guild == guild {
			worker.stop()
			delete(channels.workers, id)
		}
	}
}

func (channels *grinchChannels) wait() {
	// Wait for any refresh that started before cancellation to finish adding workers.
	channels.mu.Lock()
	channels.mu.Unlock()
	channels.wg.Wait()
}
