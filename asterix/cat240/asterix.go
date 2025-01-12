package cat240

import (
	"asterix-parser/asterix"
	"asterix-parser/pkg"

	"fmt"
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
	Cat240Specs = 0
	Field1      = 1
	Field2      = 2
	Field3      = 3
	Field4      = 4
	Field5      = 5
	Field6      = 6
	Field7      = 7
	Field8      = 8
	Field9      = 9
	Field10     = 10
	Field11     = 11
	Field12     = 12
	Field13     = 13
	Field14     = 14
	Field15     = 15
	Field16     = 16

	LenField = 17
)

var Cat240Profile = [LenField]int{
	1,            // 0 Spec (Max 2 bytes)
	2,            // 1 Data Source Identifier
	1,            // 2 Message Type
	4,            // 3 Video Record Header
	varLen,       // 4 Video Summary
	12,           // 5 Video Header Nano
	12,           // 6 Video Header Femto
	2,            // 7 Video Cells Resolution & Data Compression Indicator
	-1,           // 8 FX (Field Extension)
	5,            // 9 Video Octets & Video Cells Counters
	lowVarLen,    // 10 Video Block Low Data Volume: 1+4*n octets
	mediumVarLen, // 11 Video Block Medium Data Volume: 1+64*n octets
	highVarLen,   // 12 Video Block High Data Volume: 1+256*n octets
	3,            // 13 Time of Day
	0,            // 14 Reserved Expansion Field
	0,            // 15 SP
	-1,           // 16 FX (Field Extension)
}

func asterixFieldPut(bit int, value []byte, cat240 *asterix.Cat240) {
	switch bit {
	case Field1:
		cat240.DataSourceIdentifier.SystemAreaCode = value[0]
		cat240.DataSourceIdentifier.SystemIdentificationCode = value[1]
	case Field2:
		cat240.MessageType = value[0]
	case Field3:
		cat240.VideoRecordHeader = [4]byte{value[0], value[1], value[2], value[3]}
	case Field4:
		cat240.VideoSummary = value
	case Field5:
		cat240.VideoHeaderNano.StartAzimuth = [2]byte{value[0], value[1]}
		cat240.VideoHeaderNano.EndAzimuth = [2]byte{value[2], value[3]}
		cat240.VideoHeaderNano.Range = [4]byte{value[4], value[5], value[6], value[7]}
		cat240.VideoHeaderNano.Duration = [4]byte{value[8], value[9], value[10], value[11]}
	case Field6:
		cat240.VideoHeaderFemto.StartAzimuth = [2]byte{value[0], value[1]}
		cat240.VideoHeaderFemto.EndAzimuth = [2]byte{value[2], value[3]}
		cat240.VideoHeaderFemto.Range = [4]byte{value[4], value[5], value[6], value[7]}
		cat240.VideoHeaderFemto.Duration = [4]byte{value[8], value[9], value[10], value[11]}
	case Field7:
		cat240.DataCompressionIndicator = value[0]
		cat240.VideoCellsResolution = value[1]
	case Field8:
		// FX
	case Field9:
		cat240.VideoOctets = [2]byte{value[0], value[1]}
		cat240.VideoCells = [3]byte{value[2], value[3], value[4]}
	case Field10:
		cat240.VideoBlockLowVolume.RepetitionIndicator = value[0]
		cat240.VideoBlockLowVolume.Data = value[1:]
	case Field11:
		cat240.VideoBlockMediumVolume.RepetitionIndicator = value[0]
		cat240.VideoBlockMediumVolume.Data = value[1:]
	case Field12:
		cat240.VideoBlockHighVolume.RepetitionIndicator = value[0]
		cat240.VideoBlockHighVolume.Data = value[1:]
	case Field13:
		cat240.TimeOfDay = [3]byte{value[0], value[1], value[2]}
	case Field14:
		// Reserved Expansion
	case Field15:
		// SP
	case Field16:
		// FX
	default:
		pkg.LogError(fmt.Sprintf("Unknown field: %d", bit))
	}
}

func asterixUnpackProcess(cat240 *asterix.Cat240, raw []byte) {
	// Fspec extension
	fspecExtension := asterix.IsBitSet(raw[:1], Field8)

	// Get fspec
	len := Cat240Profile[Cat240Specs] + fspecExtension
	fspec := raw[:len]
	cat240.FieldSpecs = fspec

	// fmt.Printf("FSPEC (%d): [%x]\n", len, fspec)

	pos := 0
	for bit := Field1; bit < LenField; bit++ {
		if asterix.IsBitSet(fspec, bit) == 0 {
			continue
		}

		format := Cat240Profile[bit]
		// fmt.Printf("Bit (%d)(%d): ", bit, format)

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
				value = raw[pos:len]
			}

		} else {
			if format < 0 {
				// continue
				// fmt.Print("FX\n")
				continue
			} else {
				pos = len
				len += int(math.Abs(float64(format)))
				value = raw[pos:len]
			}
		}

		// fmt.Printf("[%x] \n", value)

		// Put the value
		asterixFieldPut(bit, value, cat240)
	}

}

func AsterixUnpackMessage(cat240 *asterix.Cat240, message *[]byte) error {
	// Check message empty
	if message == nil {
		return fmt.Errorf("cannot parse empty input")
	}
	// Parse message
	raw := *message

	// Process the header cat240 (F0)
	if raw[0] != 0xf0 {
		return fmt.Errorf("invalid cat240 header: %d", raw[0])
	}
	raw = raw[1:]

	// Process Length
	raw = raw[2:]

	// Parse the message
	asterixUnpackProcess(cat240, raw)

	return nil
}
