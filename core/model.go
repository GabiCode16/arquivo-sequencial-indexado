package core

// ProdutoRecord representa o registro de tamanho fixo em produtos.dat (97 bytes).
type ProdutoRecord struct {
	ProductID    int64
	Brand        [32]byte
	CategoryCode [48]byte
	Price        float64
	Delimiter    byte
}

// AcessoRecord representa o registro de tamanho fixo em acessos.dat (61 bytes).
type AcessoRecord struct {
	EventID   int64
	UserID    int64
	EventType [16]byte
	ProductID int64
	EventTime [20]byte
	Delimiter byte
}
