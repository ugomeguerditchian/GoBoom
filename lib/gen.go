package lib

var StatusCodeToEscape = []string{
	//"503 Too many open connections",
	"401 Unauthorized",
	"409 Conflict",
	"404 Not Found",
	"502 Bad Gateway",
	"504 Gateway Timeout",
	"407 Proxy Authentication Required",
	"400 Bad Request",
	"502 Proxy Error",
	"403 Forbidden",
	//"503 Service Unavailable",
	"504 DNS Name Not Found",
	"407 Unauthorized",
	"405 Method Not Allowed",
	//"503 Service Temporarily Unavailable",
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func RemoveDuplicates(elements []string) []string {
	// Use map to record duplicates as we find them.
	encountered := map[string]bool{}
	result := []string{}

	for v := range elements {
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}
