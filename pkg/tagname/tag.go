package tagname

// Tag - nil means any value
type Tag struct {
	Type  *string
	Value *string
}

// NewTag - format: typ:val . 'typ' and 'val' can be any string value except '*' that means any string.
