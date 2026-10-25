package core

// ProdutoRecord representa o registro de tamanho fixo em produtos.dat (97 bytes).
type ProdutoRecord struct {
	ProductID    int64
	Brand        [BrandLen]byte
	CategoryCode [CatLen]byte
	Price        float64
	Delimiter    byte
}

// AcessoRecord representa o registro de tamanho fixo em acessos.dat (61 bytes).
type AcessoRecord struct {
	EventID   int64
	UserID    int64
	EventType [TypeLen]byte
	ProductID int64
	EventTime [TimeLen]byte
	Delimiter byte
}

// ProdutoIdx representa o registro de tamanho fixo em indice_prod.idx (16 bytes).
type ProdutoIdx struct {
	ProductID int64
	Offset    int64
}

// AcessoIdx representa o registro de tamanho fixo em indice_acess.idx (16 bytes).
type AcessoIdx struct {
	EventID int64
	Offset  int64
}
