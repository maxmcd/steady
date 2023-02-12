// util is a package that should go away
package steadyutil

import (
	"crypto/rand"
	"net/http"
	"strings"
)

func ExtractAppName(r *http.Request) string {
	host := r.Host
	hostParts := strings.Split(host, ".")
	if len(hostParts) == 0 {
		return ""
	}
	return hostParts[0]
}

var XAppName = "X-App-Name"

var chars = "abcdefghijklmnopqrstuvwxyz0123456789"

// RandomString will return a random string of letters and numbers of the
// specified length. An empty string is returned if there is an error reading
// random data.
func RandomString(length int) string {
	ll := len(chars)
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	for i := 0; i < length; i++ {
		b[i] = chars[int(b[i])%ll]
	}
	return string(b)
}
