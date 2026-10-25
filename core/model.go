package core

// Tamanhos da representação binária em disco, sem o alinhamento das structs Go.
const (
	TamanhoRegistroProduto int64 = 97
	TamanhoRegistroAcesso  int64 = 61
	TamanhoEntradaIndice   int64 = 16
	RegistrosPorBloco      int64 = 1024
)

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
