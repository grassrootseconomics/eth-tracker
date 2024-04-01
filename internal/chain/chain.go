package chain

import (
	"log/slog"

	"github.com/grassrootseconomics/celoutils/v2"
)

type (
	ChainOpts struct {
		RPCEndpoint string
		TestNet     bool
		Logg        *slog.Logger
	}

	Chain struct {
		provider *celoutils.Provider
		logg     *slog.Logger
	}
)

func NewChainProvider(o ChainOpts) (*Chain, error) {
	providerOpts := celoutils.ProviderOpts{
		RpcEndpoint: o.RPCEndpoint,
		ChainId:     celoutils.MainnetChainId,
	}

	if o.TestNet {
		providerOpts.ChainId = celoutils.TestnetChainId
	}

	provider, err := celoutils.NewProvider(providerOpts)
	if err != nil {
		return nil, err
	}

	return &Chain{
		provider: provider,
		logg:     o.Logg,
	}, nil
}
