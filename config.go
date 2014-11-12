package main

import (
	"time"

	"github.com/BurntSushi/toml"
)

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) (err error) {
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

type GithubConfig struct {
	Listen string `toml:"listen"`
	Secret string `toml:"secret"`
}

type BroadcastConfig struct {
	Channel string   `toml:"channel"`
	Topic   string   `toml:"topic"`
	Timeout duration `toml:"timeout"`
}

type ArchiveConfig struct {
	ArchiveTable     string `toml:"archive_table"`
	SubscribersTable string `toml:"docker_subscribers"`
	BroadcastChannel string `toml:"broadcast-hooks"`
	HooksChannel     string `toml:"hooks_channel"`
}

type Config struct {
	Debug             bool            `toml:"debug"`
	NSQD              string          `toml:"nsqd"`
	Lookupd           string          `toml:"lookupd"`
	RethinkdbAddress  string          `toml:"rethinkdb_address"`
	RethinkdbKey      string          `toml:"rethinkdb_key"`
	RethinkdbDatabase string          `toml:"rethinkdb_database"`
	Github            GithubConfig    `toml:"github"`
	Archive           ArchiveConfig   `toml:"archive"`
	Broadcast         BroadcastConfig `toml:"broadcast"`
}

func loadConfig(path string) *Config {
	var c *Config
	if _, err := toml.DecodeFile(path, &c); err != nil {
		logger.Fatal(err)
	}
	return c
}
