package gmdbx

type Option struct {
	// Database save path
	Path       string
	Flags      EnvFlags
	Geometry   Geometry
	MaxDBS     uint16
	TxnDpLimit uint16
}

const (
	DefaultFlags = EnvSyncDurable | EnvNoTLS | EnvWriteMap | EnvLIFOReclaim | EnvNoMemInit | EnvCoalesce
	SimpleFlags  = EnvNoMetaSync | EnvSyncDurable
)

var (
	DefaultGeometry = Geometry{
		SizeLower:       1 << 30,
		SizeNow:         1 << 30,
		SizeUpper:       1 << 34,
		GrowthStep:      1 << 30,
		ShrinkThreshold: 1 << 31,
		PageSize:        1 << 16,
	}

	DefaultOption = Option{
		Path:       "db",
		Flags:      SimpleFlags,
		Geometry:   DefaultGeometry,
		MaxDBS:     1024,
		TxnDpLimit: 1024,
	}
)
