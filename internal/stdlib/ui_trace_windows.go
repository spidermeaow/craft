//go:build windows

package stdlib

import (
	"encoding/json"
	"os"
	"sort"
	"time"
)

// CRAFT_UI_TRACE=1 emits one aggregate line to stderr at session shutdown.
// Values are engine timings. Completion of draw calls is not display presentation.
// Keep only recent samples so tracing a long-running desktop app cannot grow
// memory in proportion to the number of rendered frames.
const uiTraceSampleLimit = 8192

type uiTrace struct {
	enabled             bool
	queue               []int64
	measure             []int64
	layout              []int64
	paint               []int64
	inputToEndDraw      []int64
	scrollToEndDraw     []int64
	frames              int64
	overflowNodes       int64
	visibleNodes        int64
	scrollEvents        int64
	virtualRows         int64
	rangeRequests       int64
	drawCommands        int64
	imageBytes          int64
	invalidationRegions int64
	invalidatedPixels   int64
	fullRedraws         int64
	renderer            string
	rendererFallback    string
	canvasPaintOps      int64
}

func newUITrace() uiTrace { return uiTrace{enabled: os.Getenv("CRAFT_UI_TRACE") == "1"} }

func (t *uiTrace) record(dst *[]int64, duration time.Duration) {
	if t.enabled {
		if len(*dst) == uiTraceSampleLimit {
			copy(*dst, (*dst)[uiTraceSampleLimit/2:])
			*dst = (*dst)[:uiTraceSampleLimit/2]
		}
		*dst = append(*dst, duration.Microseconds())
	}
}

func traceStats(values []int64) map[string]int64 {
	if len(values) == 0 {
		return nil
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	percentile := func(n int) int64 {
		index := (len(values)*n+99)/100 - 1
		if index < 0 {
			index = 0
		}
		return values[index]
	}
	return map[string]int64{"count": int64(len(values)), "p50_us": percentile(50), "p95_us": percentile(95), "p99_us": percentile(99)}
}

func (t *uiTrace) report() {
	if !t.enabled {
		return
	}
	payload := map[string]any{
		"kind":                        "craft-ui-engine-timing",
		"sample_limit_per_metric":     uiTraceSampleLimit,
		"unit":                        "microseconds",
		"frames":                      t.frames,
		"overflow_nodes":              t.overflowNodes,
		"visible_nodes_last_frame":    t.visibleNodes,
		"scroll_events":               t.scrollEvents,
		"virtual_rows_last_layout":    t.virtualRows,
		"visible_range_requests":      t.rangeRequests,
		"canvas_commands_last_tree":   t.drawCommands,
		"image_bytes_last_tree":       t.imageBytes,
		"invalidation_regions":        t.invalidationRegions,
		"invalidated_pixels":          t.invalidatedPixels,
		"full_redraws":                t.fullRedraws,
		"renderer":                    t.renderer,
		"renderer_fallback_reason":    t.rendererFallback,
		"canvas_paint_ops_last_frame": t.canvasPaintOps,
		"command_queue":               traceStats(t.queue),
		"measure":                     traceStats(t.measure),
		"layout":                      traceStats(t.layout),
		"paint_to_draw_complete":      traceStats(t.paint),
		"input_to_draw_complete":      traceStats(t.inputToEndDraw),
		"scroll_to_draw_complete":     traceStats(t.scrollToEndDraw),
		"note":                        "Draw completion means Direct2D EndDraw or the end of GDI draw calls, not display presentation; input correlation is the most recent delivered event followed by gui.update.",
	}
	encoded, err := json.Marshal(payload)
	if err == nil {
		_, _ = os.Stderr.Write(append(encoded, '\n'))
	}
}
