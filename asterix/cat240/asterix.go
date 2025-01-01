package cat240

import (
	"asterix-parser/asterix"
	"asterix-parser/pkg"
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"math"
)

// Variable len types
const (
	varLen       = 1000
	lowVarLen    = 1001
	mediumVarLen = 1002
	highVarLen   = 1003
)

// Data volume factors
const (
	LowDataVolumeFactor    = 4
	MediumDataVolumeFactor = 64
	HighDataVolumeFactor   = 256
)

const (
	Cat240Specs   = 0
	Cat240Field1  = 1
	Cat240Field2  = 2
	Cat240Field3  = 3
	Cat240Field4  = 4
	Cat240Field5  = 5
	Cat240Field6  = 6
	Cat240Field7  = 7
	Cat240Field8  = 8
	Cat240Field9  = 9
	Cat240Field10 = 10
	Cat240Field11 = 11
	Cat240Field12 = 12
	Cat240Field13 = 13
	Cat240Field14 = 14
	Cat240Field15 = 15
	Cat240Field16 = 16

	LenCat240Field = 17
)

var Cat240Profile = [LenCat240Field]int{
	1,            // 0 Spec (Max 2 bytes)
	2,            // 1 Data Source Identifier
	1,            // 2 Message Type
	4,            // 3 Video Record Header
	varLen,       // 4 Video Summary
	12,           // 5 Video Header Nano
	12,           // 6 Video Header Femto
	2,            // 7 Video Cells Resolution & Data Compression Indicator
	-1,           // 8 FX (Cat240Field Extension)
	5,            // 9 Video Octets & Video Cells Counters
	lowVarLen,    // 10 Video Block Low Data Volume: 1+4*n octets
	mediumVarLen, // 11 Video Block Medium Data Volume: 1+64*n octets
	highVarLen,   // 12 Video Block High Data Volume: 1+256*n octets
	3,            // 13 Time of Day
	0,            // 14 Reserved Expansion Cat240Field
	0,            // 15 SP
	-1,           // 16 FX (Cat240Field Extension)
}

func asterixFieldPut(bit int, value []byte, cat240 *asterix.Cat240) {
	switch bit {
	case Cat240Field1:
		cat240.DataSourceIdentifier.SystemAreaCode = value[0]
		cat240.DataSourceIdentifier.SystemIdentificationCode = value[1]
	case Cat240Field2:
		cat240.MessageType = value[0]
	case Cat240Field3:
		cat240.VideoRecordHeader = [4]byte{value[0], value[1], value[2], value[3]}
	case Cat240Field4:
		cat240.VideoSummary = value
	case Cat240Field5:
		cat240.VideoHeaderNano = [12]byte{value[0], value[1], value[2], value[3], value[4], value[5], value[6], value[7], value[8], value[9], value[10], value[11]}
	case Cat240Field6:
		cat240.VideoHeaderFemto = [12]byte{value[0], value[1], value[2], value[3], value[4], value[5], value[6], value[7], value[8], value[9], value[10], value[11]}
	case Cat240Field7:
		cat240.DataCompression = value[0]
		cat240.VideoCellsResolution = value[1]
	case Cat240Field8:
		// FX
	case Cat240Field9:
		cat240.VideoOctets = [2]byte{value[0], value[1]}
		cat240.VideoCells = [3]byte{value[2], value[3], value[4]}
	case Cat240Field10:
		cat240.VideoBlockLowVolume.RepetitionIndicator = value[0]
		cat240.VideoBlockLowVolume.Data = value[1:]
	case Cat240Field11:
		cat240.VideoBlockMediumVolume.RepetitionIndicator = value[0]
		cat240.VideoBlockMediumVolume.Data = value[1:]
	case Cat240Field12:
		cat240.VideoBlockHighVolume.RepetitionIndicator = value[0]
		cat240.VideoBlockHighVolume.Data = value[1:]
	case Cat240Field13:
		cat240.TimeOfDay = [3]byte{value[0], value[1], value[2]}
	case Cat240Field14:
		// Reserved Expansion
	case Cat240Field15:
		// SP
	case Cat240Field16:
		// FX
	default:
		fmt.Printf("Unknown field: %d\n", bit)
	}
}

func asterixUnpackProcess(cat240 *asterix.Cat240, raw []byte) {
	// Fspec extension
	fspecExtension := pkg.UtilBitTest(raw[:1], Cat240Field8)

	// Get fspec
	len := Cat240Profile[Cat240Specs] + fspecExtension
	fspec := raw[:len]
	cat240.FieldSpecs = fspec

	fmt.Printf("FSPEC (%d): [%x]\n", len, fspec)

	pos := 0
	for bit := Cat240Field1; bit < LenCat240Field; bit++ {
		if pkg.UtilBitTest(fspec, bit) == 0 {
			continue
		}

		format := Cat240Profile[bit]
		fmt.Printf("Bit (%d)(%d): ", bit, format)

		var value []byte
		if format == varLen || format == lowVarLen || format == mediumVarLen || format == highVarLen {
			pos = len

			// Get repetition
			len += 1
			rep := raw[pos:len]
			if format == varLen {
				len += int(rep[0])
			} else {
				factor := 1

				if format == lowVarLen {
					factor = LowDataVolumeFactor
				} else if format == mediumVarLen {
					factor = MediumDataVolumeFactor
				} else if format == highVarLen {
					factor = HighDataVolumeFactor
				}
				len += int(rep[0]) * factor

				// Compression checking
				if cat240.DataCompression>>(8-1) == 1 {
					// Compression
					blockData := raw[pos+1 : len]

					reader, err := zlib.NewReader(bytes.NewReader(blockData))
					if err != nil {
						continue
					}
					defer reader.Close()

					// Read the decompressed data from the zlib reader
					decompressedData, err := io.ReadAll(reader)
					if err != nil {
						continue
					}

					value = append(rep, decompressedData...)

				} else {
					// No compression
					value = raw[pos:len]
				}
			}

		} else {
			if format < 0 {
				// continue
				continue
			} else {
				pos = len
				len += int(math.Abs(float64(format)))
				value = raw[pos:len]
			}
		}

		fmt.Printf("[%x] \n", value)

		// Put the value
		asterixFieldPut(bit, value, cat240)
	}

}

func AsterixUnpackMessage(cat240 *asterix.Cat240, message *[]byte) {
	// Parse message
	raw := *message

	// Process the header cat240 (F0)
	raw = raw[1:]

	// Process Length
	raw = raw[2:]

	// Parse the message
	asterixUnpackProcess(cat240, raw)
}
