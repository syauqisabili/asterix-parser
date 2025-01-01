package pkg

func UtilBitTest(buf []byte, index int) int {
	index--
	for index >= 8 {
		buf = buf[1:] // Move to the next byte
		index -= 8
	}

	// Extract the bit value
	return int((buf[0] >> (7 - index)) & 0x01)
}
