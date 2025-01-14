package main

import (
	"asterix-parser/asterix"
	"asterix-parser/asterix/cat240"
	"asterix-parser/dto"
	"asterix-parser/pkg"
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"fmt"
	"io"
	"math"
)

func main() {
	// Example
	sentence := "f0016be7c80000020000b87335f239800000000003e219358004014800080052789c95d1cb4ec2401406e0c770a569b925269d96a24151316031a64a8ba0803131be890b763e830bcbc2a41b4d7c43e742efff69ebb79ccb7fe69cd95b0b2fd8d32bb620ceaf624bc1cf994e3de5195be52ccbf8de6c678ecd1eb045643e5f638f9c97b8936e9589b4c06e7646dc15e7baae87c581e3f178820d94a1c4c31cc7214e5e63eaa643198eb04b6c981824ce31dbb2acb34861f7028bb64f133d69809dc4fafd5e869dd2952b47954cc6580f38c6ec6a9664468853bc2c330cc3a2521833812e661835de153bb4b0622ebaedfbf7d226f8fafe89fc6261186eb741b0ade5f3837b97366fed8c0ed62ed3cad2b9268148a0d6f735a12150898256a02ba5cfae2929ddc29ac4ccaa87970f2af654a361d127bf5c559bfe9403252e48b549fe1e3d6be2517a4a03d3b46cf57f8e5675960fadfaa84ea77e45624ae95f142575fd0faa8aca98374432"

	// Convert hex to byte using encoding/hex
	message, err := hex.DecodeString(sentence)
	if err != nil {
		pkg.LogError(fmt.Errorf("decode string: %v", err))
		return
	}

	packet := &asterix.Cat240{}
	if err := cat240.AsterixUnpackMessage(packet, &message); err != nil {
		pkg.LogError(fmt.Errorf("parsing asterix sentence: %v", err))
		return
	}
	// Radar
	radar := &dto.Radar{}

	// Max range (Km)
	radar.MaxRange = 20

	// Get record header
	radar.Header = uint32(packet.VideoRecordHeader[0])<<24 | uint32(packet.VideoRecordHeader[1])<<16 | uint32(packet.VideoRecordHeader[2])<<8 | uint32(packet.VideoRecordHeader[3])

	// Get compression indicator
	radar.DataCompressionIndicator = packet.DataCompressionIndicator>>(8-1) == 1

	// Block data
	if asterix.IsBitSet(packet.FieldSpecs, cat240.Field10) == 1 {
		radar.VideoBlockData = packet.VideoBlockLowVolume.Data
	} else if asterix.IsBitSet(packet.FieldSpecs, cat240.Field11) == 1 {
		radar.VideoBlockData = packet.VideoBlockMediumVolume.Data
	} else if asterix.IsBitSet(packet.FieldSpecs, cat240.Field12) == 1 {
		radar.VideoBlockData = packet.VideoBlockHighVolume.Data
	}

	// Valid octet
	radar.ValidOctet = uint16(packet.VideoOctets[0])<<8 | uint16(packet.VideoOctets[1])

	// Valid cells
	radar.ValidCell = uint32(packet.VideoCells[0])<<16 | uint32(packet.VideoCells[1])<<8 | uint32(packet.VideoCells[2])

	// Check azimuth is not empty
	value := 0
	if packet.VideoHeaderNano != (asterix.VideoHeader{}) {
		// Start azimuth
		value = int(packet.VideoHeaderNano.StartAzimuth[0])<<8 | int(packet.VideoHeaderNano.StartAzimuth[1])
		radar.StartAzimuth = (360.0 / math.MaxUint16) * float32(value)
		// End azimuth
		value = int(packet.VideoHeaderNano.EndAzimuth[0])<<8 | int(packet.VideoHeaderNano.EndAzimuth[1])
		radar.EndAzimuth = (360.0 / math.MaxUint16) * float32(value)
		// Range
		radar.StartRange = uint32(packet.VideoHeaderNano.Range[0])<<24 | uint32(packet.VideoHeaderNano.Range[1])<<16 | uint32(packet.VideoHeaderNano.Range[2])<<8 | uint32(packet.VideoHeaderNano.Range[3])
		// Cell duration
		duration := uint32(packet.VideoHeaderNano.Duration[0])<<24 | uint32(packet.VideoHeaderNano.Duration[1])<<16 | uint32(packet.VideoHeaderNano.Duration[2])<<8 | uint32(packet.VideoHeaderNano.Duration[3])
		radar.CellDuration = float64(duration) * 0.000000000000001

	} else if packet.VideoHeaderFemto != (asterix.VideoHeader{}) {
		// Start azimuth
		value = int(packet.VideoHeaderFemto.StartAzimuth[0])<<8 | int(packet.VideoHeaderFemto.StartAzimuth[1])
		radar.StartAzimuth = 360.0 / math.MaxUint16 * float32(value)
		// End azimuth
		value = int(packet.VideoHeaderFemto.EndAzimuth[0])<<8 | int(packet.VideoHeaderFemto.EndAzimuth[1])
		radar.EndAzimuth = 360.0 / math.MaxUint16 * float32(value)
		// Range
		radar.StartRange = uint32(packet.VideoHeaderFemto.Range[0])<<24 | uint32(packet.VideoHeaderFemto.Range[1])<<16 | uint32(packet.VideoHeaderFemto.Range[2])<<8 | uint32(packet.VideoHeaderFemto.Range[3])
		// Cell duration
		duration := uint32(packet.VideoHeaderFemto.Duration[0])<<24 | uint32(packet.VideoHeaderFemto.Duration[1])<<16 | uint32(packet.VideoHeaderFemto.Duration[2])<<8 | uint32(packet.VideoHeaderFemto.Duration[3])
		radar.CellDuration = float64(duration) * 0.000000000000001
	}

	// Convert time of day to hh:mm:ss
	value = int(packet.TimeOfDay[0])<<16 | int(packet.TimeOfDay[1])<<8 | int(packet.TimeOfDay[2])

	secs := float32(value / 128)
	hour := int(secs / 3600)

	remaining := secs - float32((hour * 3600))
	minutes := int(remaining / 60)

	remaining -= float32(minutes * 60)
	timeOfDay := fmt.Sprintf("%02d%02d%02d", hour, minutes, int(remaining))

	radar.TimeOfDay = timeOfDay

	// Asterix data compression indicator check
	if radar.DataCompressionIndicator {
		blockData := radar.VideoBlockData

		// Decompress using zlib
		reader, err := zlib.NewReader(bytes.NewReader(blockData))
		if err != nil {
			pkg.LogError(fmt.Printf("zlib reader: %v", err))
		}
		defer reader.Close()

		// Read the decompressed data from the zlib reader
		rawData, err := io.ReadAll(reader)
		if err != nil {
			pkg.LogError(fmt.Printf("io read: %v", err))
		}

		radar.VideoBlockData = []byte{}
		radar.VideoBlockData = append(radar.VideoBlockData, rawData...) // Append the decompression
	}

	pkg.LogDebug(fmt.Printf("radar: %v", radar))
}
