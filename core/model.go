package core

// Definição das structs e regras de tamanho fixo para produtos e acessos.

// Produto representa a estrutura de dados de um produto com tamanho fixo.
type Produto struct {
	ID        uint64
	Categoria string
	Marca     string
	Preco     float64
	Ativo     byte
}

// Acesso representa a estrutura de dados de um evento de acesso com tamanho fixo.
type Acesso struct {
	ID         uint64
	UsuarioID  uint64
	ProdutoID  uint64
	TipoEvento string
	DataHora   string
	Ativo      byte
}

// EntradaIndice representa um registro no arquivo de índice parcial.
type EntradaIndice struct {
	Chave  uint64
	Offset int64
}
