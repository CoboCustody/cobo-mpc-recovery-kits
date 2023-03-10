package tss

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"math/big"

	"github.com/cobo/cobo-mpc-recovery-kits/pkg/crypto"
	"github.com/decred/dcrd/dcrec/edwards/v2"
)

type (
	Share struct {
		ID, Xi *big.Int
	}
	Shares []*Share
)

//nolint:unparam
func (shares Shares) reconstruct(curve elliptic.Curve) (*big.Int, error) {
	var secret *big.Int
	n := curve.Params().N

	shareIDs := make([]*big.Int, 0)
	for _, share := range shares {
		shareIDs = append(shareIDs, share.ID)
	}

	secret = big.NewInt(0)
	for i, share := range shares {
		t := big.NewInt(1)
		for j := 0; j < len(shareIDs); j++ {
			if j == i {
				continue
			}
			sub := new(big.Int)
			sub.Sub(shareIDs[j], share.ID)
			sub.Mod(sub, n)

			inv := new(big.Int)
			inv.ModInverse(sub, n)

			mul := new(big.Int)
			mul.Mul(shareIDs[j], inv)
			mul.Mod(mul, n)

			tMul := new(big.Int)
			tMul.Mul(t, mul)
			t = tMul.Mod(tMul, n)
		}

		sMul := new(big.Int)
		sMul.Mul(share.Xi, t)
		sMul.Mod(sMul, n)

		sAdd := new(big.Int)
		sAdd.Add(secret, sMul)
		secret = sAdd.Mod(sAdd, n)
	}
	return secret, nil
}

func (shares Shares) ReconstructECDSAKey(threshold int, curve elliptic.Curve) (*ecdsa.PrivateKey, error) {
	if shares == nil || threshold < 1 {
		return nil, fmt.Errorf("input error")
	}
	if threshold > len(shares) {
		return nil, fmt.Errorf("too little shares for threshold to reconstruct")
	}
	secret, err := shares.reconstruct(curve)
	if err != nil {
		return nil, err
	}

	return crypto.CreateECDSAPrivateKey(curve, secret), nil
}

func (shares Shares) ReconstructEDDSAKey(threshold int, curve elliptic.Curve) (*edwards.PrivateKey, error) {
	if shares == nil || threshold < 1 {
		return nil, fmt.Errorf("input error")
	}
	if threshold > len(shares) {
		return nil, fmt.Errorf("too little shares for threshold to reconstruct")
	}
	secret, err := shares.reconstruct(curve)
	if err != nil {
		return nil, err
	}

	return crypto.CreateEDDSAPrivateKey(secret)
}
