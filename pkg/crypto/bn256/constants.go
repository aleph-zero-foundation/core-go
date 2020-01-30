package bn256

import (
	"github.com/cloudflare/bn256"
)

// Order reexports cloudflare/bn256.Order.
var Order = bn256.Order

// SignatureLength is the length of the returned signatures after marshaling.
// It is not explicitly defined in the underlaying package, but it is always 64.
const SignatureLength = 64
