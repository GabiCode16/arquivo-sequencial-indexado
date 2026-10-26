package core

// Caminhos padrão dos arquivos de dados e índices.
const (
	ProdPath     = "dados/produtos.dat"
	AcessPath    = "dados/acessos.dat"
	ProdIdxPath  = "dados/indice_prod.idx"
	AcessIdxPath = "dados/indice_acess.idx"
	ExtFmt       = "%s_%d.ext"
)

// Tamanhos da representação binária em disco e dos campos de texto.
const (
	ProdSize        = 106
	AcessSize       = 70
	IdxSize         = 16
	BlockSize int64 = 1024
	BrandLen        = 32
	CatLen          = 48
	TypeLen         = 16
	TimeLen         = 20
	Delim           = '\n'
	EloNil    int64 = -1
	EloSize         = 8
	FlagSize        = 1
	DelimSize       = 1
	FlagOn    byte  = 1
	Perm            = 0600
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
	OptProd     = "1"
	OptInsProd  = "2"
	OptRemProd  = "3"
	OptInsAcess = "4"
	OptRemAcess = "5"
	OptExit     = "0"
	IDBase      = 10
	IDBits      = 64
	FloatBits   = 64
	TextPad     = " \x00"
)
