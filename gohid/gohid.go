package gohid

type HIDDevice interface {
	GetFeatureReport(b []byte) (int, error)
	SendFeatureReport(b []byte) (int, error)
	Close() error
}

func OpenHID(path string) (HIDDevice, error) {
	return openHIDInternal(path)
}
