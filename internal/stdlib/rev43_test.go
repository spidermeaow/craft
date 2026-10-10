package stdlib

import (
	"bytes"
	"context"
	rt "craft/internal/runtime"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRev43EncodingCrypto(t *testing.T) {
	ctx := context.Background()
	raw := Bytes{data: "\x00\xffCraft"}
	encoded, err := Invoke(ctx, &bytes.Buffer{}, "std.encoding.base64Encode", []any{raw})
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Invoke(ctx, &bytes.Buffer{}, "std.encoding.base64Decode", []any{encoded})
	if err != nil || decoded.(Bytes).data != raw.data {
		t.Fatalf("base64: %v %v", decoded, err)
	}
	hexValue, err := Invoke(ctx, &bytes.Buffer{}, "std.encoding.hexEncode", []any{raw})
	if err != nil {
		t.Fatal(err)
	}
	decoded, err = Invoke(ctx, &bytes.Buffer{}, "std.encoding.hexDecode", []any{hexValue})
	if err != nil || decoded.(Bytes).data != raw.data {
		t.Fatalf("hex: %v %v", decoded, err)
	}
	if _, err = Invoke(ctx, &bytes.Buffer{}, "std.encoding.hexDecode", []any{"0g"}); err == nil {
		t.Fatal("invalid hex accepted")
	}
	digest, err := Invoke(ctx, &bytes.Buffer{}, "std.crypto.sha256", []any{Bytes{data: "abc"}})
	if err != nil || len(digest.(Bytes).data) != 32 {
		t.Fatalf("sha256: %v", err)
	}
	mac, err := Invoke(ctx, &bytes.Buffer{}, "std.crypto.hmacSha256", []any{Bytes{data: "key"}, raw})
	if err != nil || len(mac.(Bytes).data) != 32 {
		t.Fatalf("hmac: %v", err)
	}
	equal, err := Invoke(ctx, &bytes.Buffer{}, "std.crypto.secureEqual", []any{mac, mac})
	if err != nil || equal != true {
		t.Fatalf("secureEqual: %v %v", equal, err)
	}
	random, err := Invoke(ctx, &bytes.Buffer{}, "std.crypto.secureRandom", []any{int64(32)})
	if err != nil || len(random.(Bytes).data) != 32 {
		t.Fatalf("random: %v", err)
	}
	uuid, err := Invoke(ctx, &bytes.Buffer{}, "std.crypto.uuid", nil)
	if err != nil || len(uuid.(string)) != 36 || uuid.(string)[14] != '4' {
		t.Fatalf("uuid: %v %v", uuid, err)
	}
}

func TestRev43URLRegex(t *testing.T) {
	ctx := context.Background()
	encoded, err := Invoke(ctx, &bytes.Buffer{}, "std.url.encode", []any{"a b+ไทย"})
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Invoke(ctx, &bytes.Buffer{}, "std.url.decode", []any{encoded})
	if err != nil || decoded != "a b+ไทย" {
		t.Fatalf("url encoding: %v %v", decoded, err)
	}
	parsed, err := Invoke(ctx, &bytes.Buffer{}, "std.url.parse", []any{"https://example.org/path?a=1#frag"})
	if err != nil || parsed.(*rt.Struct).Fields["host"] != "example.org" {
		t.Fatalf("url parse: %v %v", parsed, err)
	}
	query, err := Invoke(ctx, &bytes.Buffer{}, "std.url.query", []any{"a=1&a=2&b="})
	if err != nil || len(query.(*rt.Map).Values["a"].(*rt.Array).Items) != 2 {
		t.Fatalf("url query: %v %v", query, err)
	}
	found, err := Invoke(ctx, &bytes.Buffer{}, "std.regex.find", []any{"[0-9]+", "ไทย12"})
	if err != nil || !found.(rt.Optional).Valid {
		t.Fatalf("regex find: %v %v", found, err)
	}
	match := found.(rt.Optional).Value.(*rt.Struct)
	if match.Fields["start"] != int64(3) || match.Fields["end"] != int64(5) {
		t.Fatalf("unicode offsets: %v", match)
	}
	replaced, err := Invoke(ctx, &bytes.Buffer{}, "std.regex.replace", []any{"[0-9]+", "x12", "z"})
	if err != nil || replaced != "xz" {
		t.Fatalf("regex replace: %v %v", replaced, err)
	}
	if _, err = Invoke(ctx, &bytes.Buffer{}, "std.regex.match", []any{"(", "x"}); err == nil {
		t.Fatal("bad regex accepted")
	}
}

func TestRev43Child(t *testing.T) {
	if os.Getenv("CRAFT_REV43_CHILD") != "1" {
		return
	}
	if os.Getenv("CRAFT_REV43_CHILD_MODE") == "sleep" {
		<-time.After(30 * time.Second)
		return
	}
	_, _ = os.Stdout.WriteString("hello")
	_, _ = os.Stderr.WriteString("warning")
	os.Exit(7)
}

func TestRev43Process(t *testing.T) {
	t.Setenv("CRAFT_REV43_CHILD", "1")
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, closeScope := BeginProcessScope(context.Background())
	defer closeScope()
	session := NewSession(nil, nil)
	arguments := &rt.Array{Items: []any{"-test.run=^TestRev43Child$"}}
	result, err := session.Invoke(ctx, &bytes.Buffer{}, "std.process.run", []any{executable, arguments, int64(5000), int64(1024)})
	if err != nil {
		t.Fatal(err)
	}
	got := result.(*rt.Struct)
	if got.Fields["exitCode"] != int64(7) || got.Fields["stdout"].(Bytes).data != "hello" || got.Fields["stderr"].(Bytes).data != "warning" {
		t.Fatalf("child result: %v", got)
	}
	if _, err := session.Invoke(ctx, &bytes.Buffer{}, "std.process.run", []any{executable, arguments, int64(5000), int64(2)}); err == nil || !strings.Contains(err.Error(), "maxOutputBytes") {
		t.Fatalf("output cap: %v", err)
	}
	p, err := session.Invoke(ctx, &bytes.Buffer{}, "std.process.start", []any{executable, arguments, int64(1024)})
	if err != nil {
		t.Fatal(err)
	}
	result, err = session.Invoke(ctx, &bytes.Buffer{}, "std.process.wait", []any{p, int64(5000)})
	if err != nil || result.(*rt.Struct).Fields["exitCode"] != int64(7) {
		t.Fatalf("start/wait: %v %v", result, err)
	}
	exit, err := session.Invoke(ctx, &bytes.Buffer{}, "std.process.exitCode", []any{p})
	if err != nil || exit.(rt.Optional).Value != int64(7) {
		t.Fatalf("exit code: %v %v", exit, err)
	}
	t.Setenv("CRAFT_REV43_CHILD_MODE", "sleep")
	if _, err := session.Invoke(ctx, &bytes.Buffer{}, "std.process.run", []any{executable, arguments, int64(20), int64(1024)}); err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("timeout: %v", err)
	}
	p, err = session.Invoke(ctx, &bytes.Buffer{}, "std.process.start", []any{executable, arguments, int64(1024)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := session.Invoke(ctx, &bytes.Buffer{}, "std.process.kill", []any{p}); err != nil {
		t.Fatal(err)
	}
	if _, err := session.Invoke(ctx, &bytes.Buffer{}, "std.process.wait", []any{p, int64(5000)}); err != nil {
		t.Fatalf("wait after kill: %v", err)
	}
	ownedCtx, cleanup := BeginProcessScope(context.Background())
	p, err = session.Invoke(ownedCtx, &bytes.Buffer{}, "std.process.start", []any{executable, arguments, int64(1024)})
	if err != nil {
		t.Fatal(err)
	}
	cleanup()
	select {
	case <-p.(*Process).done:
	case <-time.After(5 * time.Second):
		t.Fatal("process scope did not clean up child")
	}
}

func TestRev43OSInformation(t *testing.T) {
	for _, name := range []string{"std.os.version", "std.os.architecture", "std.os.processId"} {
		value, err := Invoke(context.Background(), &bytes.Buffer{}, name, nil)
		if err != nil {
			t.Fatal(err)
		}
		if value == "" || value == int64(0) {
			t.Fatalf("%s returned %v", name, value)
		}
	}
	if runtime.GOOS == "windows" {
		if _, err := Invoke(context.Background(), &bytes.Buffer{}, "std.os.open", []any{"javascript:alert(1)"}); err == nil {
			t.Fatal("unsupported URL scheme accepted")
		}
	}
}
