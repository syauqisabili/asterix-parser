package asterix

type DataSourceIdentifier struct {
	SystemAreaCode           byte
	SystemIdentificationCode byte
}

type VideoBlockDataVolume struct {
	RepetitionIndicator byte
	Data                []byte
}

type Cat240 struct {
	FieldSpecs             []byte
	DataSourceIdentifier   DataSourceIdentifier
	MessageType            byte
	VideoRecordHeader      [4]byte
	VideoSummary           []byte
	VideoHeaderNano        [12]byte
	VideoHeaderFemto       [12]byte
	DataCompression        byte
	VideoCellsResolution   byte
	VideoOctets            [2]byte
	VideoCells             [3]byte
	VideoBlockLowVolume    VideoBlockDataVolume
	VideoBlockMediumVolume VideoBlockDataVolume
	VideoBlockHighVolume   VideoBlockDataVolume
	TimeOfDay              [3]byte
}
