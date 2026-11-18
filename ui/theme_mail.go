package ui

import "github.com/kivutar/goro/ui/rotheme"

// [mail] defaults: the mailbox and the read-mail window.
var (
	mailWindowW       = 450
	mailWindowH       = 366
	mailRowsPerPage   = 7
	mailReadW         = 410
	mailReadH         = 330
	mailTabW          = 72
	mailTabH          = 24
	mailComposePad    = 8
	mailSubjectMinW   = float32(150)
	mailSenderW       = float32(112)
	mailDateW         = float32(85)
	mailRowH          = float32(24)
	mailHeaderH       = float32(24)
	mailPageGap       = 8
	mailPagePad       = 8
	mailInboxGap      = 6
	mailAttachmentH   = float32(40)
	mailAttachmentGap = 8
	mailComposeGap    = 5
	mailFieldH        = float32(24)
	mailFieldGap      = 6
	mailFieldLabelW   = float32(65)
	mailReadPad       = 10
	mailReadGap       = 6
	mailSlot          = float32(40)
	mailAmountLabelH  = float32(14)
	mailAmountInset   = float32(2)
)

func init() {
	s := rotheme.Section("mail", "mailbox and read mail windows")
	s.Int("width", &mailWindowW, 280, 1400, "mailbox width")
	s.Int("height", &mailWindowH, 200, 1000, "mailbox height")
	s.Int("rows_per_page", &mailRowsPerPage, 1, 30, "inbox rows per page")
	s.Int("read_width", &mailReadW, 240, 1200, "read mail width")
	s.Int("read_height", &mailReadH, 180, 900, "read mail height")
	s.Int("tab_width", &mailTabW, 32, 240, "inbox and write tab width")
	s.Int("tab_height", &mailTabH, 14, 64, "inbox and write tab height")
	s.Int("compose_padding", &mailComposePad, 0, 32, "compose body padding")
	s.Float("subject_min_width", &mailSubjectMinW, 40, 480, "subject column minimum width")
	s.Float("sender_width", &mailSenderW, 40, 320, "sender column width")
	s.Float("date_width", &mailDateW, 40, 240, "date column width")
	s.Float("row_height", &mailRowH, 12, 64, "inbox row height")
	s.Float("header_height", &mailHeaderH, 12, 64, "inbox header height")
	s.Int("page_gap", &mailPageGap, 0, 32, "gap in the page bar")
	s.Int("page_padding", &mailPagePad, 0, 32, "page bar padding")
	s.Int("inbox_gap", &mailInboxGap, 0, 32, "gap between the inbox table and the page bar")
	s.Float("attachment_height", &mailAttachmentH, 16, 96, "attachment row height")
	s.Int("attachment_gap", &mailAttachmentGap, 0, 32, "gap beside the attachment icon")
	s.Int("compose_gap", &mailComposeGap, 0, 32, "gap between compose fields")
	s.Float("field_height", &mailFieldH, 12, 64, "compose field row height")
	s.Int("field_gap", &mailFieldGap, 0, 32, "gap between a field label and its field")
	s.Float("field_label_width", &mailFieldLabelW, 24, 200, "compose field label width")
	s.Int("read_padding", &mailReadPad, 0, 40, "read mail padding")
	s.Int("read_gap", &mailReadGap, 0, 32, "gap between read mail rows")
	s.Float("slot_size", &mailSlot, 16, 96, "attachment slot size")
	s.Float("amount_label_height", &mailAmountLabelH, 8, 32, "attachment amount label height")
	s.Float("amount_label_inset", &mailAmountInset, 0, 16, "attachment amount label inset")
}
