package tss

import (
	"gitlab.com/alephledger/core-go/pkg/crypto/bn256"
)

// CreateWTK generates a weak threshold key for the given ThresholdKeys
// i.e. a ThresholdKey which corresponds to the sum of polynomials
// which are defining the given ThresholdKeys.
// Shares may be produced only by shareProviders.
// We assume that:
//  (0) tks is a non-empty slice
//  (1) the threshold is the same for all given thresholdKeys
//  (2) the thresholdKeys were created by different processes
//  (3) the thresholdKeys have the same owner
//
// The resulting WeakThresholdKey has undefined dealer and encSKs.
func CreateWTK(tks []*ThresholdKey, shareProviders map[uint16]bool) *WeakThresholdKey {
	n := len(tks[0].vks)

	result := &WeakThresholdKey{}

	result.ThresholdKey = ThresholdKey{
		owner:     tks[0].owner,
		threshold: tks[0].threshold,
		vks:       make([]*bn256.VerificationKey, n),
	}

	result.shareProviders = shareProviders

	for _, tk := range tks {
		result.sk = bn256.AddSecretKeys(result.sk, tk.sk)
		result.globalVK = bn256.AddVerificationKeys(result.globalVK, tk.globalVK)
		for i, vk := range tk.vks {
			result.vks[i] = bn256.AddVerificationKeys(result.vks[i], vk)
		}
	}
	return result
}

// SumShares returns a share for a set of threshold keys.
// All the shares should be created by the same process.
// The given slice of shares has to be non empty.
func SumShares(shs []*Share) *Share {
	var sum *bn256.Signature
	for _, sh := range shs {
		sum = bn256.AddSignatures(sum, sh.sgn)
	}
	return &Share{
		owner: shs[0].owner,
		sgn:   sum,
	}
}
