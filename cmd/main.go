package main

import (
	"asterix-parser/asterix"
	"asterix-parser/asterix/cat240"
	"asterix-parser/domain"
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"fmt"
	"io"
	"math"
)

func main() {

	sentence := "f0016be7c80000020000b87335f239800000000003e219358004014800080052789c95d1cb4ec2401406e0c770a569b925269d96a24151316031a64a8ba0803131be890b763e830bcbc2a41b4d7c43e742efff69ebb79ccb7fe69cd95b0b2fd8d32bb620ceaf624bc1cf994e3de5195be52ccbf8de6c678ecd1eb045643e5f638f9c97b8936e9589b4c06e7646dc15e7baae87c581e3f178820d94a1c4c31cc7214e5e63eaa643198eb04b6c981824ce31dbb2acb34861f7028bb64f133d69809dc4fafd5e869dd2952b47954cc6580f38c6ec6a9664468853bc2c330cc3a2521833812e661835de153bb4b0622ebaedfbf7d226f8fafe89fc6261186eb741b0ade5f3837b97366fed8c0ed62ed3cad2b9268148a0d6f735a12150898256a02ba5cfae2929ddc29ac4ccaa87970f2af654a361d127bf5c559bfe9403252e48b549fe1e3d6be2517a4a03d3b46cf57f8e5675960fadfaa84ea77e45624ae95f142575fd0faa8aca98374432"

	// Convert hex to byte using encoding/hex
	message, err := hex.DecodeString(sentence)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	asterixCat240 := &asterix.Cat240{}
	cat240.AsterixUnpackMessage(asterixCat240, &message)

	// Radar
	radar := &domain.Radar{}

	// Compression checking
	if asterixCat240.DataCompression>>(8-1) == 1 {
		blockData := asterixCat240.VideoBlockLowVolume.Data

		reader, err := zlib.NewReader(bytes.NewReader(blockData))
		if err != nil {

		}
		defer reader.Close()

		// Read the decompressed data from the zlib reader
		rawData, err := io.ReadAll(reader)
		if err != nil {

		}

		radar.VideoBlock = rawData
	}

	value := 0
	// Check azimuth is not empty
	if asterixCat240.VideoHeaderNano != (asterix.VideoHeader{}) {
		// Start azimuth
		value = int(asterixCat240.VideoHeaderNano.StartAzimuth[0])<<8 | int(asterixCat240.VideoHeaderNano.StartAzimuth[1])
		radar.StartAzimuth = (360.0 / math.MaxUint16) * float32(value)
		// End azimuth
		value = int(asterixCat240.VideoHeaderNano.EndAzimuth[0])<<8 | int(asterixCat240.VideoHeaderNano.EndAzimuth[1])
		radar.EndAzimuth = (360.0 / math.MaxUint16) * float32(value)
		// Range
		radar.Range = uint32(asterixCat240.VideoHeaderNano.Range[0])<<24 | uint32(asterixCat240.VideoHeaderNano.Range[1])<<16 | uint32(asterixCat240.VideoHeaderNano.Range[2])<<8 | uint32(asterixCat240.VideoHeaderNano.Range[3])
		// Cell duration
		radar.CellDuration = uint32(asterixCat240.VideoHeaderNano.Duration[0])<<24 | uint32(asterixCat240.VideoHeaderNano.Duration[1])<<16 | uint32(asterixCat240.VideoHeaderNano.Duration[2])<<8 | uint32(asterixCat240.VideoHeaderNano.Duration[3])
	} else if asterixCat240.VideoHeaderFemto != (asterix.VideoHeader{}) {
		// Start azimuth
		value = int(asterixCat240.VideoHeaderFemto.StartAzimuth[0])<<8 | int(asterixCat240.VideoHeaderFemto.StartAzimuth[1])
		radar.StartAzimuth = 360.0 / math.MaxUint16 * float32(value)
		// End azimuth
		value = int(asterixCat240.VideoHeaderFemto.EndAzimuth[0])<<8 | int(asterixCat240.VideoHeaderFemto.EndAzimuth[1])
		radar.EndAzimuth = 360.0 / math.MaxUint16 * float32(value)
		// Range
		radar.Range = uint32(asterixCat240.VideoHeaderFemto.Range[0])<<24 | uint32(asterixCat240.VideoHeaderFemto.Range[1])<<16 | uint32(asterixCat240.VideoHeaderFemto.Range[2])<<8 | uint32(asterixCat240.VideoHeaderFemto.Range[3])
		// Cell duration
		radar.CellDuration = uint32(asterixCat240.VideoHeaderFemto.Duration[0])<<24 | uint32(asterixCat240.VideoHeaderFemto.Duration[1])<<16 | uint32(asterixCat240.VideoHeaderFemto.Duration[2])<<8 | uint32(asterixCat240.VideoHeaderFemto.Duration[3])
	}

	// Convert time of day to hh:mm:ss
	value = int(asterixCat240.TimeOfDay[0])<<16 | int(asterixCat240.TimeOfDay[1])<<8 | int(asterixCat240.TimeOfDay[2])

	secs := float32(value / 128)
	hour := int(secs / 3600)

	remaining := secs - float32((hour * 3600))
	minutes := int(remaining / 60)

	remaining -= float32(minutes * 60)

	timeOfDay := fmt.Sprintf("%02d%02d%02d", hour, minutes, int(remaining))

	radar.TimeOfDay = timeOfDay

	fmt.Printf("Radar: %v\n", radar)
}
