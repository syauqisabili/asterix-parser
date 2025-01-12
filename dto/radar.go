package dto

type Radar struct {
	DataCompressionIndicator bool    `json:"compression_indicator"`
	Header                   uint32  `json:"header"`
	StartAzimuth             float32 `json:"start_azimuth"` // in degrees
	EndAzimuth               float32 `json:"end_azimuth"`   // in degrees
	Range                    uint32  `json:"range"`
	CellDuration             float64 `json:"cell_duration"`
	ValidOctet               uint16  `json:"valid_octet"`
	ValidCell                uint32  `json:"valid_cell"`
	VideoBlockData           []byte  `json:"data"`
	TimeOfDay                string  `json:"time"`
}
