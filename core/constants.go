package core

// Caminhos padrão dos arquivos de dados e índices.
const (
	ProdPath     = "dados/produtos.dat"
	AcessPath    = "dados/acessos.dat"
	ProdIdxPath  = "dados/indice_prod.idx"
	AcessIdxPath = "dados/indice_acess.idx"
)

// Tamanhos da representação binária em disco e dos campos de texto.
const (
	ProdSize        = 97
	AcessSize       = 61
	IdxSize         = 16
	BlockSize int64 = 1024
	BrandLen        = 32
	CatLen          = 48
	TypeLen         = 16
	TimeLen         = 20
	Delim           = '\n'
)

// Buffer de leitura e intervalos de progresso do validador.
const (
	BufSize      = 4 * 1024 * 1024
	ProdStep     = 10_000
	AcessStep    = 1_000_000
	ProdIdxStep  = 10
	AcessIdxStep = 10_000
)

// Opções do menu, conversão de IDs e preenchimento dos campos de texto.
const (
	OptProd = "1"
	OptExit = "0"
	IDBase  = 10
	IDBits  = 64
	TextPad = " \x00"
)
