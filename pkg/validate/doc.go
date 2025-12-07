// Package validate provides validation functions for various data types and formats.
//
// Key features:
//   - Numeric validation (Lt, Gt, Lte, Gte, Positive, Negative, MultipleOf)
//   - String validation (Regex, Lowercase, Uppercase, Includes, StartsWith, EndsWith)
//   - Format validation (Email, URL, UUID, IP, CIDR, Base64, JWT)
//   - ISO datetime validation (ISODateTime, ISODate, ISOTime, ISODuration)
//   - Length and size validation (MaxLength, MinLength, MaxSize, MinSize)
//
// Usage:
//
//	if validate.Email(email) {
//	    // valid email
//	}
//
//	if validate.UUID(uuid) {
//	    // valid UUID
//	}
package validate
