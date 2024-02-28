package ipfs

type Interface interface {
	Report(data []byte) (string, error)
}
