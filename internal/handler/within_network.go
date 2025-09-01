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
	// USDGLO
	"0x4F604735c1cF31399C6E711D5962b2B3E0225AD3": true,
	// cEUR
	"0xD8763CBa276a3738E6DE85b4b3bF5FDed6D6cA73": true,
	// cREAL
	"0xe8537a3d056DA446677B9E9d6c5dB704EaAb4787": true,
	// eXOF
	"0x73F93dcc49cB8A239e2032663e9475dd5ef29A08": true,
	// PUSO
	"0x105d4A9306D2E55a71d2Eb95B81553AE1dC20d7B": true,
	// cCOP
	"0x8A567e2aE79CA692Bd748aB832081C45de4041eA": true,
	// cGHS
	"0xfAeA5F3404bbA20D3cc2f8C4B0A888F55a3c7313": true,
	// cGBP
	"0xCCF663b1fF11028f0b19058d0f7B674004a40746": true,
	// cZAR
	"0x4c35853A3B4e647fD266f4de678dCc8fEC410BF6": true,
	// cCAD
	"0xff4Ab19391af240c311c54200a492233052B6325": true,
	// cAUD
	"0x7175504C455076F15c04A2F90a8e352281F492F9": true,
	// cCHF
	"0xb55a79F398E759E43C95b979163f30eC87Ee131D": true,
	// cNGN
	"0xE2702Bd97ee33c88c8f6f92DA3B733608aa76F71": true,
	// cJPY
	"0xc45eCF20f3CD864B32D9794d6f76814aE8892e20": true,
	// axlREGEN
	"0x2E6C05f1f7D1f4Eb9A088bf12257f1647682b754": true,
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
