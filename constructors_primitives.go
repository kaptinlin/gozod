package gozod

import (
	"math/big"
	"time"

	"github.com/kaptinlin/gozod/types"
)

func String(params ...any) *ZodString[string] {
	return types.String(params...)
}

func StringPtr(params ...any) *ZodString[*string] {
	return types.StringPtr(params...)
}

func Email(params ...any) *ZodEmail[string] {
	return types.Email(params...)
}

func EmailPtr(params ...any) *ZodEmail[*string] {
	return types.EmailPtr(params...)
}

func Emoji(params ...any) *ZodEmoji[string] {
	return types.Emoji(params...)
}

func EmojiPtr(params ...any) *ZodEmoji[*string] {
	return types.EmojiPtr(params...)
}

func Base64(params ...any) *ZodBase64[string] {
	return types.Base64(params...)
}

func Base64Ptr(params ...any) *ZodBase64[*string] {
	return types.Base64Ptr(params...)
}

func Base64URL(params ...any) *ZodBase64URL[string] {
	return types.Base64URL(params...)
}

func Base64URLPtr(params ...any) *ZodBase64URL[*string] {
	return types.Base64URLPtr(params...)
}

func Hex(params ...any) *ZodHex[string] {
	return types.Hex(params...)
}

func HexPtr(params ...any) *ZodHex[*string] {
	return types.HexPtr(params...)
}

func Bool(params ...any) *ZodBool[bool] {
	return types.Bool(params...)
}

func BoolPtr(params ...any) *ZodBool[*bool] {
	return types.BoolPtr(params...)
}

func Int(params ...any) *ZodInteger[int, int] {
	return types.Int(params...)
}

func IntPtr(params ...any) *ZodInteger[int, *int] {
	return types.IntPtr(params...)
}

func Int8(params ...any) *ZodInteger[int8, int8] {
	return types.Int8(params...)
}

func Int8Ptr(params ...any) *ZodInteger[int8, *int8] {
	return types.Int8Ptr(params...)
}

func Int16(params ...any) *ZodInteger[int16, int16] {
	return types.Int16(params...)
}

func Int16Ptr(params ...any) *ZodInteger[int16, *int16] {
	return types.Int16Ptr(params...)
}

func Int32(params ...any) *ZodInteger[int32, int32] {
	return types.Int32(params...)
}

func Int32Ptr(params ...any) *ZodInteger[int32, *int32] {
	return types.Int32Ptr(params...)
}

func Int64(params ...any) *ZodInteger[int64, int64] {
	return types.Int64(params...)
}

func Int64Ptr(params ...any) *ZodInteger[int64, *int64] {
	return types.Int64Ptr(params...)
}

func Uint(params ...any) *ZodInteger[uint, uint] {
	return types.Uint(params...)
}

func UintPtr(params ...any) *ZodInteger[uint, *uint] {
	return types.UintPtr(params...)
}

func Uint8(params ...any) *ZodInteger[uint8, uint8] {
	return types.Uint8(params...)
}

func Uint8Ptr(params ...any) *ZodInteger[uint8, *uint8] {
	return types.Uint8Ptr(params...)
}

func Uint16(params ...any) *ZodInteger[uint16, uint16] {
	return types.Uint16(params...)
}

func Uint16Ptr(params ...any) *ZodInteger[uint16, *uint16] {
	return types.Uint16Ptr(params...)
}

func Uint32(params ...any) *ZodInteger[uint32, uint32] {
	return types.Uint32(params...)
}

func Uint32Ptr(params ...any) *ZodInteger[uint32, *uint32] {
	return types.Uint32Ptr(params...)
}

func Uint64(params ...any) *ZodInteger[uint64, uint64] {
	return types.Uint64(params...)
}

func Uint64Ptr(params ...any) *ZodInteger[uint64, *uint64] {
	return types.Uint64Ptr(params...)
}

func Float(params ...any) *ZodFloat[float64, float64] {
	return types.Float(params...)
}

func FloatPtr(params ...any) *ZodFloat[float64, *float64] {
	return types.FloatPtr(params...)
}

func Float32(params ...any) *ZodFloat[float32, float32] {
	return types.Float32(params...)
}

func Float32Ptr(params ...any) *ZodFloat[float32, *float32] {
	return types.Float32Ptr(params...)
}

func Float64(params ...any) *ZodFloat[float64, float64] {
	return types.Float64(params...)
}

func Float64Ptr(params ...any) *ZodFloat[float64, *float64] {
	return types.Float64Ptr(params...)
}

func Number(params ...any) *ZodFloat[float64, float64] {
	return types.Number(params...)
}

func NumberPtr(params ...any) *ZodFloat[float64, *float64] {
	return types.NumberPtr(params...)
}

func BigInt(params ...any) *types.ZodBigInt[*big.Int] {
	return types.BigInt(params...)
}

func BigIntPtr(params ...any) *types.ZodBigInt[**big.Int] {
	return types.BigIntPtr(params...)
}

func Complex(params ...any) *ZodComplex[complex128] {
	return types.Complex(params...)
}

func ComplexPtr(params ...any) *ZodComplex[*complex128] {
	return types.ComplexPtr(params...)
}

func Complex64(params ...any) *ZodComplex[complex64] {
	return types.Complex64(params...)
}

func Complex64Ptr(params ...any) *ZodComplex[*complex64] {
	return types.Complex64Ptr(params...)
}

func Complex128(params ...any) *ZodComplex[complex128] {
	return types.Complex128(params...)
}

func Complex128Ptr(params ...any) *ZodComplex[*complex128] {
	return types.Complex128Ptr(params...)
}

func Time(params ...any) *ZodTime[time.Time] {
	return types.Time(params...)
}

func TimePtr(params ...any) *ZodTime[*time.Time] {
	return types.TimePtr(params...)
}

func IPv4(params ...any) *ZodIPv4[string] {
	return types.IPv4(params...)
}

func IPv4Ptr(params ...any) *ZodIPv4[*string] {
	return types.IPv4Ptr(params...)
}

func IPv6(params ...any) *ZodIPv6[string] {
	return types.IPv6(params...)
}

func IPv6Ptr(params ...any) *ZodIPv6[*string] {
	return types.IPv6Ptr(params...)
}

func CIDRv4(params ...any) *ZodCIDRv4[string] {
	return types.CIDRv4(params...)
}

func CIDRv4Ptr(params ...any) *ZodCIDRv4[*string] {
	return types.CIDRv4Ptr(params...)
}

func CIDRv6(params ...any) *ZodCIDRv6[string] {
	return types.CIDRv6(params...)
}

func CIDRv6Ptr(params ...any) *ZodCIDRv6[*string] {
	return types.CIDRv6Ptr(params...)
}

func URL(params ...any) *ZodURL[string] {
	return types.URL(params...)
}

func URLPtr(params ...any) *ZodURL[*string] {
	return types.URLPtr(params...)
}

func Hostname(params ...any) *ZodHostname[string] {
	return types.Hostname(params...)
}

func HostnamePtr(params ...any) *ZodHostname[*string] {
	return types.HostnamePtr(params...)
}

func MAC(params ...any) *ZodMAC[string] {
	return types.MAC(params...)
}

func MACPtr(params ...any) *ZodMAC[*string] {
	return types.MACPtr(params...)
}

func MACWithDelimiter(delimiter string, params ...any) *ZodMAC[string] {
	return types.MACWithDelimiter(delimiter, params...)
}

func E164(params ...any) *ZodE164[string] {
	return types.E164(params...)
}

func E164Ptr(params ...any) *ZodE164[*string] {
	return types.E164Ptr(params...)
}

func HTTPURL(params ...any) *ZodURL[string] {
	return types.HTTPURL(params...)
}

func HTTPURLPtr(params ...any) *ZodURL[*string] {
	return types.HTTPURLPtr(params...)
}

func Iso(params ...any) *ZodIso[string] {
	return types.Iso(params...)
}

func IsoPtr(params ...any) *ZodIso[*string] {
	return types.IsoPtr(params...)
}

func IsoDateTime(params ...any) *ZodIso[string] {
	return types.IsoDateTime(params...)
}

func IsoDateTimePtr(params ...any) *ZodIso[*string] {
	return types.IsoDateTimePtr(params...)
}

func IsoDate(params ...any) *ZodIso[string] {
	return types.IsoDate(params...)
}

func IsoDatePtr(params ...any) *ZodIso[*string] {
	return types.IsoDatePtr(params...)
}

func IsoTime(params ...any) *ZodIso[string] {
	return types.IsoTime(params...)
}

func IsoTimePtr(params ...any) *ZodIso[*string] {
	return types.IsoTimePtr(params...)
}

func IsoDuration(params ...any) *ZodIso[string] {
	return types.IsoDuration(params...)
}

func IsoDurationPtr(params ...any) *ZodIso[*string] {
	return types.IsoDurationPtr(params...)
}

func CUID(params ...any) *ZodCUID[string] {
	return types.CUID(params...)
}

func CUIDPtr(params ...any) *ZodCUID[*string] {
	return types.CUIDPtr(params...)
}

func CUID2(params ...any) *ZodCUID2[string] {
	return types.CUID2(params...)
}

func CUID2Ptr(params ...any) *ZodCUID2[*string] {
	return types.CUID2Ptr(params...)
}

func GUID(params ...any) *ZodGUID[string] {
	return types.GUID(params...)
}

func GUIDPtr(params ...any) *ZodGUID[*string] {
	return types.GUIDPtr(params...)
}

func ULID(params ...any) *ZodULID[string] {
	return types.ULID(params...)
}

func ULIDPtr(params ...any) *ZodULID[*string] {
	return types.ULIDPtr(params...)
}

func XID(params ...any) *ZodXID[string] {
	return types.XID(params...)
}

func XIDPtr(params ...any) *ZodXID[*string] {
	return types.XIDPtr(params...)
}

func KSUID(params ...any) *ZodKSUID[string] {
	return types.KSUID(params...)
}

func KSUIDPtr(params ...any) *ZodKSUID[*string] {
	return types.KSUIDPtr(params...)
}

func NanoID(params ...any) *ZodNanoID[string] {
	return types.NanoID(params...)
}

func NanoIDPtr(params ...any) *ZodNanoID[*string] {
	return types.NanoIDPtr(params...)
}

func UUID(params ...any) *ZodUUID[string] {
	return types.UUID(params...)
}

func UUIDPtr(params ...any) *ZodUUID[*string] {
	return types.UUIDPtr(params...)
}

func UUIDv4(params ...any) *ZodUUID[string] {
	return types.UUIDv4(params...)
}

func UUIDv4Ptr(params ...any) *ZodUUID[*string] {
	return types.UUIDv4Ptr(params...)
}

func UUIDv6(params ...any) *ZodUUID[string] {
	return types.UUIDv6(params...)
}

func UUIDv6Ptr(params ...any) *ZodUUID[*string] {
	return types.UUIDv6Ptr(params...)
}

func UUIDv7(params ...any) *ZodUUID[string] {
	return types.UUIDv7(params...)
}

func UUIDv7Ptr(params ...any) *ZodUUID[*string] {
	return types.UUIDv7Ptr(params...)
}

func JWT(params ...any) *ZodJWT[string] {
	return types.JWT(params...)
}

func JWTPtr(params ...any) *ZodJWT[*string] {
	return types.JWTPtr(params...)
}

var (
	PrecisionMinute      = types.PrecisionMinute
	PrecisionSecond      = types.PrecisionSecond
	PrecisionDecisecond  = types.PrecisionDecisecond
	PrecisionCentisecond = types.PrecisionCentisecond
	PrecisionMillisecond = types.PrecisionMillisecond
	PrecisionMicrosecond = types.PrecisionMicrosecond
	PrecisionNanosecond  = types.PrecisionNanosecond
)
