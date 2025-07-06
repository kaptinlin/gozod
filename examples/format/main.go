package main

import (
	"fmt"
	"regexp"

	"github.com/kaptinlin/gozod"
)

// Showcases built-in format validators: Email, URL, ISO date/time, JWT.
func main() {
	fmt.Println("Format validation examples")

	// Email (Unicode allowed).
	email := gozod.Email().Unicode()
	for _, e := range []string{"user@example.com", "用户@例子.com", "bad"} {
		if _, err := email.Parse(e); err != nil {
			fmt.Printf("✗ email %q -> %v\n", e, err)
		} else {
			fmt.Printf("✓ email %q\n", e)
		}
	}

	// URL restricted to HTTPS protocol.
	httpsURL := gozod.URL(gozod.URLOptions{Protocol: regexp.MustCompile(`^https$`)})
	for _, u := range []string{"https://example.com", "http://insecure.com"} {
		if _, err := httpsURL.Parse(u); err != nil {
			fmt.Printf("✗ url %s -> %v\n", u, err)
		} else {
			fmt.Printf("✓ url %s\n", u)
		}
	}

	// ISO-8601 date string.
	isoDate := gozod.IsoDate()
	for _, d := range []string{"2023-12-25", "12/25/2023"} {
		if _, err := isoDate.Parse(d); err != nil {
			fmt.Printf("✗ date %s -> %v\n", d, err)
		} else {
			fmt.Printf("✓ date %s\n", d)
		}
	}

	// JWT structure (any alg).
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NSJ9.signature" //nolint:gosec
	if _, err := gozod.JWT().Parse(token); err != nil {
		fmt.Println("✗ jwt ->", err)
	} else {
		fmt.Println("✓ jwt valid structure")
	}
}
