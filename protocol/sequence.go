package protocol

var (
	seq = 885511
)

func getSeq() int {
	seq++
	return seq
}
