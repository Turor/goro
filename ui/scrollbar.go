package ui

func scrollbarSafeIntWidth(width int) int {
	if width <= ROScrollbarGutter {
		return 0
	}
	return width - ROScrollbarGutter
}
