package ddns

type DDns interface {
	UpsertRecord(value string, ip string) error
}
