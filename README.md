# Arquitetura do Sistema: Arquivo Sequencial-Indexado

Projeto para a disciplina de Algoritmos e Estruturas de Dados II, com foco na organização de dados sequencial-indexada utilizando arquivos binários em memória secundária e índices parciais.

## Estrutura do Projeto

```text
├── main.go               # Ponto de entrada, menu interativo e orquestração do sistema
├── core/
│   ├── model.go          # Definição das structs de dados e representações em memória
│   ├── storage.go        # Manipulação de arquivos binários (.dat), operações de I/O e Seek
│   ├── index.go          # Gerenciamento, busca e manutenção dos índices parciais (.idx)
│   └── queries.go        # Implementação das consultas analíticas sobre a base de dados
├── dados/
│   ├── produtos.dat      # Arquivo de dados de produtos (registros de tamanho fixo)
│   ├── acessos.dat       # Arquivo de dados de acessos (registros de tamanho fixo)
│   ├── indice_prod.idx   # Arquivo de índice parcial para produtos
│   └── indice_acess.idx  # Arquivo de índice parcial para acessos
├── go.mod                # Gerenciador de módulos Go
└── README.md             # Especificação e documentação da arquitetura
```

## Como Executar

### Pré-requisitos
- [Go](https://go.dev/) instalado (versão 1.20 ou superior).

### Execução direta
Na raiz do projeto, execute:

```bash
go run .
```

### Compilação do binário
Para gerar o executável:

```bash
# Windows
go build -o app.exe
./app.exe

# Linux / macOS
go build -o app
./app
```
