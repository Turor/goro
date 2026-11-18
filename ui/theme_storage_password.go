package ui

import "github.com/kivutar/goro/ui/rotheme"

// Storage password prompt.
var (
	storagePasswordWindowW  = 326
	storagePasswordContentH = 116
	storagePasswordLabelW   = 104
	storagePasswordFieldH   = 24
	storagePasswordMessageH = 30
	storagePasswordPadX     = 14
	storagePasswordPadY     = 12
	storagePasswordGap      = 8
	storagePasswordFieldGap = 8
	storagePasswordWindowH  int
)

func recomputeStoragePasswordMetrics() {
	storagePasswordWindowH = ROWindowTitleHeight + storagePasswordContentH + ROWindowFooterHeight
}

func init() {
	s := rotheme.Section("storage_password", "storage password dialog")
	s.Int("width", &storagePasswordWindowW, 200, 800, "dialog width")
	s.Int("content_height", &storagePasswordContentH, 48, 320, "dialog content height")
	s.Int("label_width", &storagePasswordLabelW, 40, 240, "password label width")
	s.Int("field_height", &storagePasswordFieldH, 14, 48, "password field height")
	s.Int("message_height", &storagePasswordMessageH, 12, 80, "status message height")
	s.Int("padding_x", &storagePasswordPadX, 0, 48, "content horizontal padding")
	s.Int("padding_y", &storagePasswordPadY, 0, 48, "content vertical padding")
	s.Int("gap", &storagePasswordGap, 0, 32, "gap between content rows")
	s.Int("field_gap", &storagePasswordFieldGap, 0, 32, "gap between a label and its field")
	rotheme.OnLayoutApplied(recomputeStoragePasswordMetrics)
	recomputeStoragePasswordMetrics()
}
