package regex

import (
	"regexp"
	"testing"
)

// matchCase is a table-driven test case for regex matching.
type matchCase = struct {
	input string
	want  bool
}

// runCases runs subtests for each case in the table.
func runCases(t *testing.T, re *regexp.Regexp, cases []matchCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Helper()
			if got := re.MatchString(tc.input); got != tc.want {
				t.Errorf("MatchString(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestEmail(t *testing.T) {
	runCases(t, Email, []matchCase{
		{"user@example.com", true},
		{"user+tag@example.com", true},
		{"user@sub.domain.com", true},
		{"a@b.co", true},
		{"first.last@example.com", true},
		{"user'name@example.com", true},
		{"user-name@example.com", true},
		{"user_name@example.com", true},
		{"", false},
		{"@example.com", false},
		{"user@", false},
		{"user@domain", false},
		{".user@example.com", false},
		{"user.@example.com", false},
		{"user..name@example.com", false},
		{"user @example.com", false},
		{"user@.example.com", false},
		{"user@example..com", false},
		{"user@-example.com", false},
	})
}

func TestHTML5Email(t *testing.T) {
	runCases(t, HTML5Email, []matchCase{
		{"test@example.com", true},
		{"user.name+tag@example.co.uk", true},
		{"x@x.au", true},
		{"user!def@example.com", true},
		{"user#hash@example.com", true},
		{"user%percent@example.com", true},
		{"", false},
		{"plainaddress", false},
		{"@missing-local.org", false},
	})
}

func TestRFC5322Email(t *testing.T) {
	runCases(t, RFC5322Email, []matchCase{
		{"user@example.com", true},
		{"user@[192.168.1.1]", true},
		{`"quoted user"@example.com`, true},
		{"user.name@example.co.uk", true},
		{"", false},
		{"plainaddress", false},
		{"@missing-local.org", false},
	})
}

func TestUnicodeEmail(t *testing.T) {
	runCases(t, UnicodeEmail, []matchCase{
		{"user@example.com", true},
		{"用户@example.com", true},
		{"user@例え.jp", true},
		{"", false},
		{"@example.com", false},
		{"user@", false},
	})
}

func TestBrowserEmail(t *testing.T) {
	runCases(t, BrowserEmail, []matchCase{
		{"test@example.com", true},
		{"user.name+tag@example.co.uk", true},
		{"", false},
		{"plainaddress", false},
	})
}

func TestURL(t *testing.T) {
	runCases(t, URL, []matchCase{
		{"http://example.com", true},
		{"https://example.com/path", true},
		{"ftp://files.example.com", true},
		{"ws://socket.example.com", true},
		{"", false},
		{"not-a-url", false},
	})
}

func TestHTTPURL(t *testing.T) {
	runCases(t, HTTPURL, []matchCase{
		{"http://example.com", true},
		{"https://example.com/path?q=1", true},
		{"", false},
		{"ftp://files.example.com", false},
	})
}

func TestIPv4(t *testing.T) {
	runCases(t, IPv4, []matchCase{
		{"0.0.0.0", true},
		{"127.0.0.1", true},
		{"255.255.255.255", true},
		{"192.168.1.1", true},
		{"", false},
		{"256.0.0.0", false},
		{"1.2.3", false},
		{"1.2.3.4.5", false},
	})
}

func TestIPv6(t *testing.T) {
	runCases(t, IPv6, []matchCase{
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},
		{"::1", true},
		{"fe80::1", true},
		{"", false},
	})
}

func TestHostname(t *testing.T) {
	runCases(t, Hostname, []matchCase{
		{"example.com", true},
		{"sub.example.com", true},
		{"localhost", true},
		{"", false},
		{"-invalid.com", false},
	})
}

func TestDomain(t *testing.T) {
	runCases(t, Domain, []matchCase{
		{"example.com", true},
		{"sub.example.com", true},
		{"example.co.uk", true},
		{"", false},
		{"localhost", false},
	})
}

func TestMAC(t *testing.T) {
	t.Run("colon", func(t *testing.T) {
		runCases(t, MACDefault, []matchCase{
			{"00:1A:2B:3C:4D:5E", true},
			{"00:1a:2b:3c:4d:5e", true},
			{"", false},
			{"00:1A:2B:3C:4D", false},
			{"GG:HH:II:JJ:KK:LL", false},
		})
	})

	t.Run("dash", func(t *testing.T) {
		dash := MAC("-")
		runCases(t, dash, []matchCase{
			{"00-1A-2B-3C-4D-5E", true},
			{"00-1a-2b-3c-4d-5e", true},
			{"00:1A:2B:3C:4D:5E", false},
		})
	})
}

func TestUUID(t *testing.T) {
	runCases(t, UUID, []matchCase{
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"00000000-0000-0000-0000-000000000000", true},
		{"", false},
		{"not-a-uuid", false},
		{"550e8400-e29b-01d4-a716-446655440000", false},
	})
}

func TestUUID4(t *testing.T) {
	runCases(t, UUID4, []matchCase{
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"", false},
		{"550e8400-e29b-51d4-a716-446655440000", false},
	})
}

func TestGUID(t *testing.T) {
	runCases(t, GUID, []matchCase{
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"550e8400-e29b-01d4-0716-446655440000", true},
		{"", false},
	})
}

func TestCUID(t *testing.T) {
	runCases(t, CUID, []matchCase{
		{"cjld2cyuq0000t3rmniod1foy", true},
		{"Cjld2cyuq0000t3rmniod1foy", true},
		{"", false},
		{"ajld2cyuq0000t3rmniod1foy", false},
	})
}

func TestULID(t *testing.T) {
	runCases(t, ULID, []matchCase{
		{"01ARZ3NDEKTSV4RRFFQ69G5FAV", true},
		{"", false},
		{"short", false},
	})
}

func TestNanoID(t *testing.T) {
	runCases(t, NanoID, []matchCase{
		{"V1StGXR8_Z5jdHi6B-myT", true},
		{"", false},
		{"short", false},
	})
}

func TestBase64(t *testing.T) {
	runCases(t, Base64, []matchCase{
		{"", true},
		{"SGVsbG8gV29ybGQ=", true},
		{"dGVzdA==", true},
		{"!!!invalid!!!", false},
	})
}

func TestBase64URL(t *testing.T) {
	runCases(t, Base64URL, []matchCase{
		{"SGVsbG8gV29ybGQ", true},
		{"dGVzdA==", true},
		{"abc_def-ghi", true},
		{"invalid chars!", false},
	})
}

func TestDate(t *testing.T) {
	runCases(t, Date, []matchCase{
		{"2024-01-15", true},
		{"2024-02-29", true},
		{"2023-02-29", false},
		{"", false},
		{"2024-13-01", false},
	})
}

func TestDefaultTime(t *testing.T) {
	runCases(t, DefaultTime, []matchCase{
		{"12:30", true},
		{"12:30:45", true},
		{"12:30:45.123", true},
		{"", false},
		{"25:00", false},
	})
}

func TestDefaultDatetime(t *testing.T) {
	runCases(t, DefaultDatetime, []matchCase{
		{"2024-01-15T12:30:45Z", true},
		{"2024-01-15T12:30:45+05:30", true},
		{"", false},
		{"2024-01-15", false},
	})
}

func TestDuration(t *testing.T) {
	runCases(t, Duration, []matchCase{
		{"P1Y", true},
		{"P1Y2M3D", true},
		{"PT1H30M", true},
		{"P1W", true},
		{"", false},
	})
}

func TestE164(t *testing.T) {
	runCases(t, E164, []matchCase{
		{"+14155552671", true},
		{"+442071234567", true},
		{"", false},
		{"14155552671", false},
		{"+1", false},
	})
}

func TestEmoji(t *testing.T) {
	runCases(t, Emoji, []matchCase{
		{"😀", true},
		{"🚀", true},
		{"", false},
		{"hello", false},
	})
}

func TestStringRegex(t *testing.T) {
	t.Run("bounded", func(t *testing.T) {
		re := StringRegex(2, 5)
		runCases(t, re, []matchCase{
			{"ab", true},
			{"abcde", true},
			{"a", false},
			{"abcdef", false},
		})
	})

	t.Run("unbounded", func(t *testing.T) {
		re := StringRegex(3, 0)
		runCases(t, re, []matchCase{
			{"abc", true},
			{"abcdefghij", true},
			{"ab", false},
		})
	})
}

func TestStringRegexCache(t *testing.T) {
	re1 := StringRegex(1, 10)
	re2 := StringRegex(1, 10)
	if re1 != re2 {
		t.Error("StringRegex(1, 10) returned different instances, want cached")
	}
}
