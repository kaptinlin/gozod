package regex

import "regexp"

var (
	// Email provides practical email validation that rejects common invalid patterns.
	// Rejects IP-address domains, leading/trailing/consecutive dots, and domains without TLD.
	Email = regexp.MustCompile(`^[A-Za-z0-9_'+\-]+([A-Za-z0-9_'+\-]*\.[A-Za-z0-9_'+\-]+)*@[A-Za-z0-9]([A-Za-z0-9\-]*[A-Za-z0-9])?(\.[A-Za-z0-9]([A-Za-z0-9\-]*[A-Za-z0-9])?)*\.[A-Za-z]{2,}$`)

	// HTML5Email matches HTML5 input[type=email] validation per the WHATWG spec.
	HTML5Email = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

	// RFC5322Email matches RFC 5322-compliant email addresses.
	RFC5322Email = regexp.MustCompile(`^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`)

	// UnicodeEmail matches emails allowing Unicode characters with length limits.
	UnicodeEmail = regexp.MustCompile(`^[^\s@"]{1,64}@[^\s@]{1,255}$`)

	// BrowserEmail is an alias for [HTML5Email] (identical pattern).
	BrowserEmail = HTML5Email
)
