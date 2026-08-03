package util

func TruncateString[String ~string | ~[]byte](s String, n int) String {
	if n >= len(s) {
		return s
	}

	// Back up until we find the beginning of a UTF-8 encoding.
	for n > 0 && s[n-1]&0xc0 == 0x80 { // 0x10... is a continuation byte
		n--
	}

	// If we're at the beginning of a multi-byte encoding, back up one more to
	// skip it. It's possible the value was already complete, but it's simpler
	// if we only have to check in one direction.
	//
	// Otherwise, we have a single-byte code (0x00... or 0x01...).
	if n > 0 && s[n-1]&0xc0 == 0xc0 { // 0x11... starts a multibyte encoding
		n--
	}
	return s[:n]
}
