package core

// ProdutoRecord representa um produto de 106 bytes no arquivo principal ou extensão.
type ProdutoRecord struct {
	ProductID    int64
	Brand        [BrandLen]byte
	CategoryCode [CatLen]byte
	Price        float64
	Elo          int64 // Offset na extensão do bloco; EloNil encerra a cadeia.
	Excluido     bool
	Delimiter    byte
}

// AcessoRecord representa um acesso de 70 bytes no arquivo principal ou extensão.
type AcessoRecord struct {
	EventID   int64
	UserID    int64
	EventType [TypeLen]byte
	ProductID int64
	EventTime [TimeLen]byte
	Elo       int64
	Excluido  bool
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
