package stdlib

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// UiDraw is immutable paint data. Coordinates are local logical pixels in a Canvas.
type UiDraw struct {
	kind                 string
	target               string
	x, y, w, h           int64
	color                uint32 // straight ARGB
	radius, thickness    int64
	text                 string
	textUTF16            []uint16
	fontSize, fontWeight int64
	textAlign            int64
	image                *uiImage
	tabStop, disabled    bool
	accessibleName       string
}

func cloneDraw(d UiDraw) UiDraw {
	d.textUTF16 = append([]uint16(nil), d.textUTF16...)
	return d
}

func sameDraws(a, b []UiDraw) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		x, y := a[i], b[i]
		if x.kind != y.kind || x.target != y.target || x.x != y.x || x.y != y.y || x.w != y.w || x.h != y.h || x.color != y.color || x.radius != y.radius || x.thickness != y.thickness || x.text != y.text || x.fontSize != y.fontSize || x.fontWeight != y.fontWeight || x.textAlign != y.textAlign || x.image != y.image || x.tabStop != y.tabStop || x.disabled != y.disabled || x.accessibleName != y.accessibleName {
			return false
		}
	}
	return true
}

func canvasMetrics(root *UiNode) (int64, int64) {
	seen := map[*uiImage]bool{}
	var count, bytes int64
	var walk func(*UiNode)
	walk = func(n *UiNode) {
		if n == nil {
			return
		}
		count += int64(len(n.draws))
		for _, d := range n.draws {
			if d.image != nil && !seen[d.image] {
				seen[d.image] = true
				bytes += int64(len(d.image.pixels))
			}
		}
		for _, c := range n.children {
			walk(c)
		}
	}
	walk(root)
	return count, bytes
}

func (d *UiDraw) String() string { return fmt.Sprintf("UiDraw(%s)", d.kind) }

// Pixels are premultiplied BGRA, accepted by Direct2D and AlphaBlend.
type uiImage struct {
	path          string
	width, height int
	pixels        []byte
}

type uiCachedImage struct {
	asset    *uiImage
	modTime  int64
	fileSize int64
}

const maxCanvasCommands = 2048
const maxImageDimension = 2048
const maxImageFileBytes = 16 << 20
const maxSessionImageBytes = 64 << 20

func validCanvasColor(value int64) bool { return value >= 0 && value <= 0xffffffff }

func validDrawBounds(x, y, w, h int64) bool {
	return x >= -8192 && x <= 8192 && y >= -8192 && y <= 8192 && w > 0 && w <= 8192 && h > 0 && h <= 8192 && w*h <= 4<<20
}

func (s *Session) decodeUIImage(path string) (*uiImage, error) {
	if path == "" || len(path) > 2048 || strings.ContainsRune(path, 0) || !utf8.ValidString(path) {
		return nil, fmt.Errorf("UiAssetError: invalid image path")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("UiAssetError: image path: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("UiAssetError: image %q: %w", path, err)
	}
	if !info.Mode().IsRegular() || info.Size() > maxImageFileBytes {
		return nil, fmt.Errorf("UiAssetError: image must be a regular file of at most 16 MiB")
	}
	if cached, ok := s.uiImages[abs]; ok && cached.modTime == info.ModTime().UnixNano() && cached.fileSize == info.Size() {
		return cached.asset, nil
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, fmt.Errorf("UiAssetError: read image: %w", err)
	}
	if len(data) > maxImageFileBytes {
		return nil, fmt.Errorf("UiAssetError: image file grew beyond 16 MiB")
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "png" && format != "jpeg") {
		return nil, fmt.Errorf("UiAssetError: image must be PNG or JPEG")
	}
	if config.Width < 1 || config.Height < 1 || config.Width > maxImageDimension || config.Height > maxImageDimension {
		return nil, fmt.Errorf("UiAssetError: image dimensions exceed 2048x2048")
	}
	needed := int64(config.Width) * int64(config.Height) * 4
	if s.uiImageBytes+needed > maxSessionImageBytes {
		return nil, fmt.Errorf("UiAssetError: image cache exceeds 64 MiB")
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("UiAssetError: decode image: %w", err)
	}
	asset := &uiImage{path: abs, width: config.Width, height: config.Height, pixels: make([]byte, needed)}
	bounds := decoded.Bounds()
	for y := 0; y < config.Height; y++ {
		for x := 0; x < config.Width; x++ {
			r, g, b, a := decoded.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			i := (y*config.Width + x) * 4
			asset.pixels[i], asset.pixels[i+1], asset.pixels[i+2], asset.pixels[i+3] = byte(b>>8), byte(g>>8), byte(r>>8), byte(a>>8)
		}
	}
	if s.uiImages == nil {
		s.uiImages = make(map[string]uiCachedImage)
	}
	s.uiImages[abs] = uiCachedImage{asset: asset, modTime: info.ModTime().UnixNano(), fileSize: info.Size()}
	s.uiImageBytes += needed
	return asset, nil
}

func (s *Session) makeUIDraw(name string, args []any) (*UiDraw, error) {
	d := &UiDraw{}
	switch name {
	case "std.ui.hitRegion":
		d.target = args[0].(string)
		d.x, d.y, d.w, d.h = args[1].(int64), args[2].(int64), args[3].(int64), args[4].(int64)
		if d.target == "" || len(d.target) > 128 || !utf8.ValidString(d.target) || strings.ContainsRune(d.target, 0) || !validDrawBounds(d.x, d.y, d.w, d.h) {
			return nil, fmt.Errorf("UiDrawError: invalid hit region")
		}
		d.kind = "hitRegion"
		d.tabStop = true
	case "std.ui.fillRect", "std.ui.strokeRect":
		d.x, d.y, d.w, d.h = args[0].(int64), args[1].(int64), args[2].(int64), args[3].(int64)
		color := args[4].(int64)
		if !validDrawBounds(d.x, d.y, d.w, d.h) || !validCanvasColor(color) {
			return nil, fmt.Errorf("UiDrawError: invalid rectangle or color")
		}
		d.color = uint32(color)
		if name == "std.ui.fillRect" {
			d.kind = "fillRect"
			d.radius = args[5].(int64)
		} else {
			d.kind = "strokeRect"
			d.thickness = args[5].(int64)
			d.radius = args[6].(int64)
		}
		if d.radius < 0 || d.radius > 4096 || d.thickness < 0 || d.thickness > 256 {
			return nil, fmt.Errorf("UiDrawError: invalid radius or stroke width")
		}
		if d.kind == "strokeRect" && d.thickness == 0 {
			return nil, fmt.Errorf("UiDrawError: stroke width must be positive")
		}
	case "std.ui.drawLine":
		x1, y1, x2, y2 := args[0].(int64), args[1].(int64), args[2].(int64), args[3].(int64)
		color, thickness := args[4].(int64), args[5].(int64)
		if x1 < -8192 || x1 > 8192 || x2 < -8192 || x2 > 8192 || y1 < -8192 || y1 > 8192 || y2 < -8192 || y2 > 8192 || !validCanvasColor(color) || thickness < 1 || thickness > 256 {
			return nil, fmt.Errorf("UiDrawError: invalid line")
		}
		d.kind = "line"
		d.x = x1
		d.y = y1
		d.w = x2
		d.h = y2
		d.color = uint32(color)
		d.thickness = thickness
	case "std.ui.drawText", "std.ui.drawTextStyled":
		d.text = args[0].(string)
		d.x, d.y, d.w, d.h = args[1].(int64), args[2].(int64), args[3].(int64), args[4].(int64)
		color := args[5].(int64)
		if len(d.text) > 4096 || !utf8.ValidString(d.text) || strings.ContainsRune(d.text, 0) || !validDrawBounds(d.x, d.y, d.w, d.h) || !validCanvasColor(color) {
			return nil, fmt.Errorf("UiDrawError: invalid text command")
		}
		d.kind = "text"
		d.color = uint32(color)
		d.textUTF16 = utf16.Encode([]rune(d.text))
		if name == "std.ui.drawTextStyled" {
			d.fontSize, d.fontWeight = args[6].(int64), args[7].(int64)
			if d.fontSize < 8 || d.fontSize > 96 || d.fontWeight < 100 || d.fontWeight > 900 || d.fontWeight%100 != 0 {
				return nil, fmt.Errorf("UiDrawError: font size must be 8..96 and weight 100..900 in steps of 100")
			}
			switch args[8].(string) {
			case "start":
				d.textAlign = 0
			case "center":
				d.textAlign = 1
			case "end":
				d.textAlign = 2
			default:
				return nil, fmt.Errorf("UiDrawError: text alignment must be start, center or end")
			}
		}
	case "std.ui.drawImage":
		d.x, d.y, d.w, d.h = args[1].(int64), args[2].(int64), args[3].(int64), args[4].(int64)
		if !validDrawBounds(d.x, d.y, d.w, d.h) {
			return nil, fmt.Errorf("UiDrawError: invalid image bounds")
		}
		asset, err := s.decodeUIImage(args[0].(string))
		if err != nil {
			return nil, err
		}
		d.kind = "image"
		d.image = asset
	default:
		return nil, fmt.Errorf("UiCapabilityError: unknown draw command")
	}
	return d, nil
}
