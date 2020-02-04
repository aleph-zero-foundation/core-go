// Package crypto specifies our assumptions about the trustworthiness of processes.
// Since less than 1/3 of them can be byzantine we can define some trust thresholds.
package crypto

// MinimalQuorum is the minimal possible size of a subset forming a quorum within nProcesses.
func MinimalQuorum(nProcesses uint16) uint16 {
	return nProcesses - nProcesses/3
}

// MinimalTrusted is the minimal size of a subset of nProcesses, that guarantees
// that the subset contains at least one honest process.
func MinimalTrusted(nProcesses uint16) uint16 {
	return (nProcesses-1)/3 + 1
}
