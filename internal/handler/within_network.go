package handler

import "context"

// Add whatever is the in:
// Sarafu Registry AND (Celo Allowed Gas Token OR Known Busy Token e.g. USDGLO)
var busyContracts = map[string]bool{
	// cUSD
	"0x765DE816845861e75A25fCA122bb6898B8B1282a": true,
	// USDT
	"0x48065fbBE25f71C9282ddf5e1cD6D6A887483D5e": true,
	// cKES
	"0x456a3D042C0DbD3db53D5489e98dFb038553B0d0": true,
	// USDC
	"0xcebA9300f2b948710d2653dD7B07f33A8B32118C": true,
}

func (hc *HandlerContainer) checkWithinNetwork(ctx context.Context, contractAddress string, from string, to string) (bool, error) {
	if !busyContracts[contractAddress] {
		return true, nil
	}

	exists, err := hc.cache.ExistsNetwork(ctx, contractAddress, from, to)
	if err != nil {
		return false, err
	}

	return exists, nil
}
