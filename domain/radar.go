package domain

type Radar struct {
	StartAzimuth float32
	EndAzimuth   float32
	Range        uint32
	CellDuration uint32
	VideoBlock   []byte
	TimeOfDay    string
}
