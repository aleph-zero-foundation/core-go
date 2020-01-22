// Package crypto specifies our assumptions about the trustworthiness of processes.
// Sincce less than 1/3 of them can be byzantine we can define some trust thresholds.
package crypto

// MinimalQuorum is the minimal possible size of a subset forming a quorum within nProcesses.
func MinimalQuorum(nProcesses int) int {
	return nProcesses - nProcesses/3
}

// MinimalTrusted is the minimal size of a subset of nProcesses, that guarantees
// that the subset contains at least one honest process.
func MinimalTrusted(nProcesses int) int {
	return (nProcesses-1)/3 + 1
}
