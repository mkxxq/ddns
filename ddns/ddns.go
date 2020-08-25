package ddns

type Record struct {
	Value string
	ID    string
	RR    string
	Type  string
}

type DDns interface {
	UpdateRecord(value string, r *Record) error
	GetRecord(subDomain string) (*Record, error)
}
