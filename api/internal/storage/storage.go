package storage

type Interface interface {
	GetBalance(recordId int) (int, error)
}
