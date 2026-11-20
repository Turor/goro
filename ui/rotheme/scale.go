package rotheme

// contentScale is the uniform scale of the window currently laying out or
// drawing. 1 means the player has not resized that window. DrawText and
// custom widgets read it so labels and cells shrink with the frame.
var contentScale float32 = 1

// ContentScale returns the scale for the window whose layout or draw is
// running. Callers outside that pass see 1.
func ContentScale() float32 {
	if contentScale <= 0 {
		return 1
	}
	return contentScale
}

// SetContentScale installs the scale for the window pass. A non-positive
// value is treated as 1 so a missed setup cannot collapse every widget.
func SetContentScale(scale float32) {
	if scale <= 0 {
		scale = 1
	}
	contentScale = scale
}
