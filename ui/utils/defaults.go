package utils

type defaults struct {
	MaxW     int
	MaxH     int
	InputMax int
	W        int
	H        int
	BorderX  int
	BorderY  int
	PadX     int
	PadY     int
}

var DefaultStruct defaults = defaults{
	MaxW:     1000,
	MaxH:     300,
	InputMax: 100,
	W:        200,
	H:        50,
	PadX:     1,
	PadY:     0,
}
