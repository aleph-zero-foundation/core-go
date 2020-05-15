package rmcbox

// Status represents a state of a reliable multicast.
type Status byte

const (
	// Unknown means we either never saw any data related to an instance of RMC, or we deleted it.
	Unknown Status = iota

	// Data means we received or sent data that is being multicast, but we have not yet signed it.
	Data

	// Signed means we signed the piece of data, but we have not yet received the proof that it has been multicast successfully.
	Signed

	// Finished means we received a proof that the data has been multicast successfully.
	Finished
)
