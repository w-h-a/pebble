package idgen

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"strconv"
	"time"
)

var (
	timeNow = time.Now
)

func Generate(prefix, title, description string, nonce int) string {
	input := title + "|" + description + "|" + strconv.FormatInt(timeNow().UnixNano(), 10) + "|" + strconv.Itoa(nonce)

	hash := sha256.Sum256([]byte(input))

	mod := new(big.Int).SetInt64(1_679_616)
	n := new(big.Int).SetBytes(hash[:3])
	n.Mod(n, mod)
	suffix := n.Text(36)

	for len(suffix) < 4 {
		suffix = "0" + suffix
	}

	return fmt.Sprintf("%s-%s", prefix, suffix)
}
