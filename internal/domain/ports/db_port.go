package ports

type DbConfig interface {
	CreateTable() error
}
