package stdlib

import (
	"context"
	"craft/internal/ast"
	rt "craft/internal/runtime"
	"crypto/hmac"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

const rev43MaxPattern = 4096
const rev43MaxText = 16 << 20

func init() {
	add := func(name string, result ast.Type, params ...ast.Type) {
		signatures[name] = Signature{Params: params, Result: result}
	}
	add("std.encoding.base64Encode", ast.String, ast.Bytes)
	add("std.encoding.base64Decode", ast.Bytes, ast.String)
	add("std.encoding.hexEncode", ast.String, ast.Bytes)
	add("std.encoding.hexDecode", ast.Bytes, ast.String)
	add("std.crypto.sha256", ast.Bytes, ast.Bytes)
	add("std.crypto.hmacSha256", ast.Bytes, ast.Bytes, ast.Bytes)
	add("std.crypto.secureRandom", ast.Bytes, ast.Int)
	add("std.crypto.uuid", ast.String)
	add("std.crypto.secureEqual", ast.Bool, ast.Bytes, ast.Bytes)
	add("std.url.parse", "UrlParts", ast.String)
	add("std.url.encode", ast.String, ast.String)
	add("std.url.decode", ast.String, ast.String)
	add("std.url.query", ast.String.Array().Map(), ast.String)
	add("std.regex.match", ast.Bool, ast.String, ast.String)
	add("std.regex.find", ast.Type("RegexMatch").Optional(), ast.String, ast.String)
	add("std.regex.replace", ast.String, ast.String, ast.String, ast.String)
	add("std.regex.split", ast.String.Array(), ast.String, ast.String)
}

func rev43DataStructs() []*ast.Struct {
	makeStruct := func(name string, names []string, types []ast.Type) *ast.Struct {
		s := &ast.Struct{Name: name}
		for i, n := range names {
			s.Fields = append(s.Fields, ast.Field{Name: n, Type: types[i]})
		}
		return s
	}
	return []*ast.Struct{
		makeStruct("UrlParts", []string{"scheme", "host", "path", "rawQuery", "fragment"},
			[]ast.Type{ast.String, ast.String, ast.String, ast.String, ast.String}),
		makeStruct("RegexMatch", []string{"text", "start", "end"},
			[]ast.Type{ast.String, ast.Int, ast.Int}),
	}
}

func rev43DataError(kind, message string) error { return toolingError(kind, message) }

func rev43DataInvoke(ctx context.Context, name string, a []any) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	switch name {
	case "std.encoding.base64Encode":
		return base64.StdEncoding.EncodeToString([]byte(a[0].(Bytes).data)), nil
	case "std.encoding.hexEncode":
		return hex.EncodeToString([]byte(a[0].(Bytes).data)), nil
	case "std.encoding.base64Decode":
		input := a[0].(string)
		if len(input) > base64.StdEncoding.EncodedLen(MaxBodyBytes) {
			return nil, rev43DataError("EncodingLimitError", "base64 input exceeds limit")
		}
		result, err := base64.StdEncoding.Strict().DecodeString(input)
		if err != nil {
			return nil, rev43DataError("EncodingError", "invalid standard Base64")
		}
		if len(result) > MaxBodyBytes {
			return nil, rev43DataError("EncodingLimitError", "decoded Bytes exceed 16 MiB")
		}
		return Bytes{data: string(result)}, nil
	case "std.encoding.hexDecode":
		input := a[0].(string)
		if len(input) > 2*MaxBodyBytes {
			return nil, rev43DataError("EncodingLimitError", "hex input exceeds limit")
		}
		result, err := hex.DecodeString(input)
		if err != nil {
			return nil, rev43DataError("EncodingError", "invalid hexadecimal input")
		}
		return Bytes{data: string(result)}, nil
	case "std.crypto.sha256":
		sum := sha256.Sum256([]byte(a[0].(Bytes).data))
		return Bytes{data: string(sum[:])}, nil
	case "std.crypto.hmacSha256":
		mac := hmac.New(sha256.New, []byte(a[0].(Bytes).data))
		_, _ = mac.Write([]byte(a[1].(Bytes).data))
		return Bytes{data: string(mac.Sum(nil))}, nil
	case "std.crypto.secureEqual":
		return hmac.Equal([]byte(a[0].(Bytes).data), []byte(a[1].(Bytes).data)), nil
	case "std.crypto.secureRandom":
		size := a[0].(int64)
		if size < 0 || size > 1<<20 {
			return nil, rev43DataError("CryptoLimitError", "secureRandom length must be 0..1048576")
		}
		data := make([]byte, size)
		if _, err := cryptorand.Read(data); err != nil {
			return nil, rev43DataError("CryptoError", "secure random source failed")
		}
		return Bytes{data: string(data)}, nil
	case "std.crypto.uuid":
		data := make([]byte, 16)
		if _, err := cryptorand.Read(data); err != nil {
			return nil, rev43DataError("CryptoError", "secure random source failed")
		}
		data[6] = (data[6] & 0x0f) | 0x40
		data[8] = (data[8] & 0x3f) | 0x80
		s := hex.EncodeToString(data)
		return fmt.Sprintf("%s-%s-%s-%s-%s", s[:8], s[8:12], s[12:16], s[16:20], s[20:]), nil
	case "std.url.parse":
		input := a[0].(string)
		if len(input) > 1<<20 || !utf8.ValidString(input) {
			return nil, rev43DataError("UrlError", "URL is too large or invalid UTF-8")
		}
		parsed, err := url.Parse(input)
		if err != nil {
			return nil, rev43DataError("UrlError", "invalid URL")
		}
		return nativeStruct("UrlParts", parsed.Scheme, parsed.Host, parsed.Path, parsed.RawQuery, parsed.Fragment), nil
	case "std.url.encode":
		input := a[0].(string)
		if len(input) > 1<<20 || !utf8.ValidString(input) {
			return nil, rev43DataError("UrlError", "query component is too large or invalid UTF-8")
		}
		return url.QueryEscape(input), nil
	case "std.url.decode":
		input := a[0].(string)
		if len(input) > 3<<20 {
			return nil, rev43DataError("UrlError", "query component is too large")
		}
		decoded, err := url.QueryUnescape(input)
		if err != nil || !utf8.ValidString(decoded) || len(decoded) > 1<<20 {
			return nil, rev43DataError("UrlError", "invalid query component")
		}
		return decoded, nil
	case "std.url.query":
		input := a[0].(string)
		if len(input) > 1<<20 {
			return nil, rev43DataError("UrlError", "query exceeds 1 MiB")
		}
		values, err := url.ParseQuery(strings.TrimPrefix(input, "?"))
		if err != nil {
			return nil, rev43DataError("UrlError", "invalid URL query")
		}
		for key, parts := range values {
			if !utf8.ValidString(key) {
				return nil, rev43DataError("UrlError", "invalid query UTF-8")
			}
			for _, part := range parts {
				if !utf8.ValidString(part) {
					return nil, rev43DataError("UrlError", "invalid query UTF-8")
				}
			}
		}
		return multiMap(values, false), nil
	}
	if strings.HasPrefix(name, "std.regex.") {
		pattern, input := a[0].(string), a[1].(string)
		if len(pattern) > rev43MaxPattern || len(input) > rev43MaxText || !utf8.ValidString(pattern) || !utf8.ValidString(input) {
			return nil, rev43DataError("RegexLimitError", "regex pattern or input exceeds limit or has invalid UTF-8")
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, rev43DataError("RegexError", "invalid RE2 pattern")
		}
		switch name {
		case "std.regex.match":
			return re.MatchString(input), nil
		case "std.regex.find":
			loc := re.FindStringIndex(input)
			if loc == nil {
				return rt.Optional{}, nil
			}
			return rt.Some(nativeStruct("RegexMatch", input[loc[0]:loc[1]], int64(utf8.RuneCountInString(input[:loc[0]])), int64(utf8.RuneCountInString(input[:loc[1]])))), nil
		case "std.regex.replace":
			replacement := a[2].(string)
			if len(replacement) > rev43MaxText || !utf8.ValidString(replacement) {
				return nil, rev43DataError("RegexLimitError", "replacement exceeds limit or has invalid UTF-8")
			}
			locations := re.FindAllStringSubmatchIndex(input, 100001)
			if len(locations) > 100000 {
				return nil, rev43DataError("RegexLimitError", "regex replace exceeds 100000 matches")
			}
			var result strings.Builder
			previous := 0
			for _, loc := range locations {
				piece := re.ExpandString(nil, replacement, input, loc)
				if result.Len()+loc[0]-previous+len(piece) > rev43MaxText {
					return nil, rev43DataError("RegexLimitError", "regex result exceeds 16 MiB")
				}
				result.WriteString(input[previous:loc[0]])
				result.Write(piece)
				previous = loc[1]
			}
			if result.Len()+len(input)-previous > rev43MaxText {
				return nil, rev43DataError("RegexLimitError", "regex result exceeds 16 MiB")
			}
			result.WriteString(input[previous:])
			return result.String(), nil
		case "std.regex.split":
			parts := re.Split(input, 100001)
			if len(parts) > 100000 {
				return nil, rev43DataError("RegexLimitError", "regex split exceeds 100000 parts")
			}
			values := &rt.Array{Items: make([]any, len(parts))}
			for i, part := range parts {
				values.Items[i] = part
			}
			return values, nil
		}
	}
	return nil, rev43DataError("RuntimeError", "unknown standard-library operation")
}
