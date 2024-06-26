package config

import (
	"github.com/rarimo/zkverifier-kit/identity"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/kit/pgdb"
)

type Config struct {
	comfig.Logger
	pgdb.Databaser
	comfig.Listenerer
	identity.VerifierProvider
	Broadcasterer
	AirdropConfiger
	PriceAPIConfiger

	airdrop  comfig.Once
	verifier comfig.Once
	routing  comfig.Once
	getter   kv.Getter
}

func New(getter kv.Getter) *Config {
	return &Config{
		getter:           getter,
		Databaser:        pgdb.NewDatabaser(getter),
		Listenerer:       comfig.NewListenerer(getter),
		Logger:           comfig.NewLogger(getter, comfig.LoggerOpts{}),
		VerifierProvider: identity.NewVerifierProvider(getter),
		Broadcasterer:    NewBroadcaster(getter),
		AirdropConfiger:  NewAirdropConfiger(getter),
		PriceAPIConfiger: NewPriceAPIConfiger(getter),
	}
}
