// util is a package that should go away
package steadyutil

import (
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
