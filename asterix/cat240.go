package asterix

type DataSourceIdentifier struct {
	SystemAreaCode           byte
	SystemIdentificationCode byte
}

type VideoBlockDataVolume struct {
	RepetitionIndicator byte
	Data                []byte
}

type VideoHeader struct {
	StartAzimuth [2]byte
	EndAzimuth   [2]byte
	Range        [4]byte
	Duration     [4]byte
}

type Cat240 struct {
	FieldSpecs               []byte
	DataSourceIdentifier     DataSourceIdentifier
	MessageType              byte
	VideoRecordHeader        [4]byte
	VideoSummary             []byte
	VideoHeaderNano          VideoHeader
	VideoHeaderFemto         VideoHeader
	DataCompressionIndicator byte
	VideoCellsResolution     byte
	VideoOctets              [2]byte
	VideoCells               [3]byte
	VideoBlockLowVolume      VideoBlockDataVolume
	VideoBlockMediumVolume   VideoBlockDataVolume
	VideoBlockHighVolume     VideoBlockDataVolume
	TimeOfDay                [3]byte
}