# Arquivo Sequencial-Indexado

Projeto da disciplina de Algoritmos e Estruturas de Dados II para estudar a organização sequencial-indexada em memória secundária. Eventos de comércio eletrônico são convertidos de CSV para arquivos binários com registros de tamanho fixo, acompanhados de índices parciais que armazenam chaves e posições em bytes.

O pré-processamento é feito em Python. As rotinas de leitura, validação e geração de índices são implementadas em Go, no módulo `asi`.

## Estado atual

| Recurso | Situação |
| --- | --- |
| Conversão de CSV em produtos e acessos binários | Implementada em `scripts/formatarDados.py` |
| Modelos binários, leitura e posicionamento | Implementados em `core/model.go` e `core/storage.go` |
| Validação dos arquivos `.dat` e `.idx` | Implementada em `core/validation.go`, utilizada por `cmd/validador` |
| Geração de índices parciais a cada 1.024 registros | Implementada em `core/index_generation.go`, utilizada por `cmd/indices` |
| Consulta de produto por ID | Implementada em `core.BuscarProdutoPorID`, com busca binária no índice e leitura do bloco em `produtos.dat` |
| Menu da aplicação | Implementado em `main.go` para consultar produto por ID e sair |
| Busca de acesso, demais consultas e manutenção dos índices no pacote `core` | Pendentes |
| Consultas de produto por ID | Implementada na aplicação principal |
| Consultas de acesso e análises | Pendentes; `core/queries.go` contém apenas a declaração do pacote e um comentário |

## Estrutura do projeto

```text
.
├── cmd/
│   ├── indices/
│   │   └── main.go          # Geração dos índices parciais
│   └── validador/
│       └── main.go          # Validação sequencial dos arquivos .dat
├── core/
│   ├── model.go            # Estruturas dos registros de dados e índices
│   ├── storage.go          # Abertura, leitura binária, tamanho e Seek
│   ├── paths.go            # Caminhos padrão dos executáveis
│   ├── records.go          # Formatos e regras compartilhadas dos registros
│   ├── index.go            # Consulta de produtos pelo índice parcial
│   ├── index_generation.go # Geração compartilhada dos índices
│   ├── validation.go       # Validação compartilhada de dados e índices
│   ├── files_test.go       # Testes de geração, validação e busca
│   └── queries.go          # Reservado para consultas
├── scripts/
│   └── formatarDados.py    # Limpeza e conversão do CSV
├── .gitignore
├── go.mod
├── main.go
└── README.md
```

As pastas `csvs/` e `dados/` são locais e estão no `.gitignore`. O CSV e os arquivos gerados não acompanham o repositório:

```text
csvs/
└── 2019-Oct.csv            # Entrada, fornecida pelo usuário
dados/                     # Pasta criada pelo script Python
├── produtos.dat           # Gerado pelo script Python
├── acessos.dat            # Gerado pelo script Python
├── indice_prod.idx        # Gerado pelo comando de índices
└── indice_acess.idx        # Gerado pelo comando de índices
```

## Pré-requisitos

- Python 3 com suporte a f-strings (3.6 ou superior). O script utiliza apenas `csv`, `os`, `struct` e `sys`, da biblioteca padrão; não há dependências para instalar nem arquivo `requirements.txt`.
- Uma instalação do Go capaz de atender à versão `1.27.2`, declarada no `go.mod`. Confira essa exigência ao preparar o ambiente; a indicação anterior de Go 1.20 não corresponde ao módulo atual.
- O arquivo `2019-Oct.csv`, do dataset de comportamento de comércio eletrônico do Kaggle citado no script.
- Espaço em disco para o CSV, os dois arquivos de dados e seus índices. O consumo de memória do processamento de produtos cresce com a quantidade de produtos distintos.

Execute os comandos a seguir **na raiz do projeto**, pois os caminhos de entrada e saída são relativos ao diretório de execução.

## Preparação e execução

### 1. Disponibilizar o CSV

Crie a pasta `csvs/` e coloque nela o arquivo com o nome `2019-Oct.csv`. No PowerShell:

```powershell
New-Item -ItemType Directory -Force csvs
```

No Linux ou macOS:

```sh
mkdir -p csvs
```

O CSV deve estar em UTF-8, usar vírgula como separador e conter as seguintes colunas no cabeçalho, em qualquer ordem:

```csv
event_time,event_type,product_id,category_code,brand,price,user_id
2019-10-01 00:00:00 UTC,view,1001,electronics.smartphone,marca_exemplo,199.90,5001
```

Colunas adicionais são ignoradas. Para usar outros caminhos, altere `CAMINHO_CSV`, `CAMINHO_PRODUTOS_DAT` e `CAMINHO_ACESSOS_DAT` em `scripts/formatarDados.py`.

### 2. Gerar os arquivos binários

```sh
python scripts/formatarDados.py
```

O script cria `dados/` automaticamente e percorre o CSV duas vezes: primeiro para gerar produtos e depois para gerar acessos. Os arquivos `.dat` de destino são sobrescritos. Ao final de cada etapa, exibe o tamanho do arquivo, a quantidade de registros e o resto da divisão pelo tamanho do registro.

### 3. Gerar os índices parciais

```sh
go run ./cmd/indices
```

O comando lê os arquivos `.dat` e cria os índices que ainda não existem, informando o resultado de cada operação. Retorna código de saída 1 quando ocorre uma falha. A geração usa um arquivo temporário e só o renomeia para o destino após concluir e fechar a escrita.

**Após gerar novamente os dados, remova os índices antigos antes de executar esse comando.** Índices existentes são ignorados, mesmo que estejam vazios ou desatualizados.

No PowerShell:

```powershell
Remove-Item -LiteralPath dados/indice_prod.idx, dados/indice_acess.idx -ErrorAction SilentlyContinue
go run ./cmd/indices
```

No Linux ou macOS:

```sh
rm -f dados/indice_prod.idx dados/indice_acess.idx
go run ./cmd/indices
```

### 4. Validar os dados e índices

```sh
go run ./cmd/validador
```

O validador percorre os quatro arquivos e verifica:

- Tamanho total compatível com os registros: 97 bytes para produtos, 61 para acessos e 16 para índices.
- Presença de pelo menos um registro em cada arquivo.
- IDs em ordem estritamente crescente, sem duplicatas.
- Delimitador `\n` nos registros de dados.
- Offsets dos índices correspondentes ao início de cada bloco de 1.024 registros.
- Quantidade de registros lidos compatível com o tamanho do arquivo e término em EOF.

As leituras usam buffer de 4 MiB. O comando retorna código de saída 1 se algum arquivo falhar. A validação dos índices é estrutural; ela não compara suas chaves com os arquivos de dados. A consulta de produto faz essa comparação nas entradas que acessa. Referências entre acessos e produtos e o conteúdo semântico dos campos não são validados.

### Aplicação principal e compilação

```sh
go run ./cmd/indices
go run .
```

O primeiro comando cria os índices que ainda não existem. O segundo abre o menu da aplicação, onde já é possível consultar um produto pelo ID e ver sua marca, categoria e preço. Se os arquivos de dados forem gerados novamente, remova e regenere os índices antes de iniciar a aplicação.

Para compilar os utilitários no Windows:

```sh
go build -o validador.exe ./cmd/validador
go build -o indices.exe ./cmd/indices
```

No Linux ou macOS, use nomes de saída sem a extensão `.exe`. Execute os binários a partir da raiz para preservar os caminhos relativos dos dados.

## Regras de preparação dos dados

### Produtos

- São deduplicados por `product_id` em um dicionário em memória e ordenados por essa chave antes da gravação.
- IDs que não podem ser convertidos para inteiro são descartados.
- O preço da primeira ocorrência é preservado; preços vazios ou que falham na conversão são substituídos por `0.0`.
- Marca e categoria inicialmente vazias podem ser preenchidas por ocorrências posteriores do mesmo produto.
- Textos passam por remoção de espaços nas extremidades, truncamento e preenchimento à direita.

### Acessos

- Cada linha aceita gera um evento, mantendo a ordem de entrada do CSV.
- O `event_id` é criado pelo script como uma sequência a partir de 1; ele não vem do CSV e não representa uma ordenação por data.
- Linhas com `user_id` ou `product_id` que falham na conversão para inteiro são descartadas.
- O texto ` UTC` é removido de `event_time`; não há conversão de fuso horário nem validação do formato da data.
- Os registros são acumulados em lotes de 100.000 antes da escrita, sem manter todos os acessos em memória.

Os textos são codificados em ASCII com descarte de caracteres não ASCII. O preenchimento é feito com espaços antes da codificação; se caracteres forem descartados, `struct.pack` completa os bytes restantes com `\x00`. Textos com acentos podem perder caracteres e conter bytes nulos no final do campo.

O script espera um CSV com cabeçalho e linhas bem formadas. Não há tratamento completo para cabeçalhos ausentes, linhas incompletas ou inteiros fora do intervalo de `int64`.

## Formatos binários

Todos os arquivos usam **little-endian**, sem cabeçalho. Os inteiros são `int64` com sinal e os preços são `float64` IEEE 754. Os offsets abaixo são relativos ao início de cada registro.

Os tamanhos correspondem à serialização com `struct.pack` em Python e `encoding/binary` em Go. O tamanho de uma struct em memória em Go pode incluir alinhamento e não deve ser usado para calcular posições em disco.

### `produtos.dat`: 97 bytes por registro

Formato Python: `<q32s48sd1s`. Estrutura Go: `core.ProdutoRecord`.

| Campo | Tipo Go | Offset | Bytes |
| --- | --- | ---: | ---: |
| `ProductID` | `int64` | 0 | 8 |
| `Brand` | `[32]byte` | 8 | 32 |
| `CategoryCode` | `[48]byte` | 40 | 48 |
| `Price` | `float64` | 88 | 8 |
| `Delimiter` | `byte` (`0x0A`) | 96 | 1 |

### `acessos.dat`: 61 bytes por registro

Formato Python: `<qq16sq20s1s`. Estrutura Go: `core.AcessoRecord`.

| Campo | Tipo Go | Offset | Bytes |
| --- | --- | ---: | ---: |
| `EventID` | `int64` | 0 | 8 |
| `UserID` | `int64` | 8 | 8 |
| `EventType` | `[16]byte` | 16 | 16 |
| `ProductID` | `int64` | 32 | 8 |
| `EventTime` | `[20]byte` | 40 | 20 |
| `Delimiter` | `byte` (`0x0A`) | 60 | 1 |

Tipos de evento esperados incluem `view`, `cart` e `purchase`, mas o script não restringe os valores a essa lista. Uma data como `2019-10-01 00:00:00` ocupa 19 caracteres e recebe um espaço para completar o campo.

O byte final `\n` faz parte do registro binário; os arquivos devem ser lidos pelo tamanho fixo dos registros, pois outros campos também podem conter o byte `0x0A`.

### Índices: 16 bytes por entrada

| Arquivo | Estrutura Go | Bytes 0–7 | Bytes 8–15 |
| --- | --- | --- | --- |
| `indice_prod.idx` | `core.ProdutoIdx` | `ProductID` (`int64`) | `Offset` (`int64`) |
| `indice_acess.idx` | `core.AcessoIdx` | `EventID` (`int64`) | `Offset` (`int64`) |

Os índices não possuem delimitador. Cada entrada referencia o primeiro registro de um bloco de **1.024 registros**, incluindo o primeiro bloco do arquivo. A constante `core.RegistrosPorBloco`, compartilhada pela geração, validação e consulta, representa uma quantidade de registros, não de bytes.

Para a entrada de índice `k`, contando a partir de zero:

```text
offset em produtos.dat = k × 1024 × 97
offset em acessos.dat  = k × 1024 × 61
quantidade de entradas para N registros = teto(N / 1024)
tamanho do índice = quantidade de entradas × 16 bytes
```

Por exemplo, 1.025 registros geram duas entradas: uma para o registro de posição 0 e outra para o de posição 1.024. O offset de qualquer registro de posição `i` é `i × tamanho_do_registro`.

A consulta de produto localiza por busca binária a entrada de índice cuja chave seja a maior que não ultrapasse a chave procurada, posiciona `produtos.dat` com `Seek` e percorre até 1.024 registros do bloco. As entradas consultadas também são comparadas com seus registros de dados para detectar índices desatualizados ou inconsistentes.

## Verificação e próximos passos

Para compilar e verificar os pacotes com o mecanismo de testes do Go:

```sh
go test ./...
```

Os testes usam arquivos temporários e cobrem geração e reconstrução de índices, buscas entre blocos, IDs ausentes, dados e índices inválidos e preservação do índice anterior em caso de falha. A validação da base local deve ser executada separadamente com `go run ./cmd/validador`.

As consultas previstas são:

- Obter a marca e o preço de um produto pelo ID.
- Obter a data/hora e o tipo de um evento pelo ID.
- Obter a marca do produto associado a um evento e a data desse evento.

As consultas de acesso e de relacionamento entre evento e produto ainda precisam de busca indexada, implementação no pacote `core` e integração à aplicação principal.
