package trafficshaper

const (
	Bps int64 = 1 << (10 * iota)
	KBps
	MBps
	GBps

	Unlimited int64 = 1<<63 - 1
)
