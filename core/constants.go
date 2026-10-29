package core

// Caminhos padrão dos arquivos de dados e índices.
const (
	ProdPath     string = "dados/produtos.dat"
	AcessPath    string = "dados/acessos.dat"
	ProdIdxPath  string = "dados/indice_prod.idx"
	AcessIdxPath string = "dados/indice_acess.idx"
)

// Tamanhos da representação binária em disco e dos campos de texto.
const (
	ProdSize  int64 = 106
	AcessSize int64 = 70
	IdxSize   int64 = 16
	BlockSize int64 = 1024
	BrandLen  int64 = 32
	CatLen    int64 = 48
	TypeLen   int64 = 16
	TimeLen   int64 = 20
	EloSize   int64 = 8
	Perm      int64 = 0600
)

// Buffer de leitura e intervalos de progresso do validador.
const (
	BufSize      int64 = 4 * 1024 * 1024
	ProdStep     int64 = 10_000
	AcessStep    int64 = 1_000_000
	ProdIdxStep  int64 = 10
	AcessIdxStep int64 = 10_000
)
