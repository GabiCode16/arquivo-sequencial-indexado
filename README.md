# Arquivo Sequencial-Indexado

Projeto da disciplina de Algoritmos e Estruturas de Dados II para estudar a organização sequencial-indexada em memória secundária. Eventos de comércio eletrônico são convertidos de CSV para arquivos binários com registros de tamanho fixo, acompanhados de índices parciais que armazenam chaves e posições em bytes.

O pré-processamento é feito em Python. As rotinas de leitura, validação e geração de índices são implementadas em Go, no módulo `asi`.

## Estado atual

| Recurso | Situação |
| --- | --- |
| Conversão de CSV em produtos e acessos binários | Implementada em `scripts/formatarDados.py` |
| Modelos binários, leitura e posicionamento | Implementados em `core/model.go` e `core/storage.go` |
| Validação dos arquivos `.dat` | Implementada em `cmd/validador` |
| Geração de índices parciais a cada 1.024 registros | Implementada em `cmd/indices` |
| Consultas por ID | Implementadas com busca pelo índice parcial e com busca binária direta nos arquivos `.dat` |
| Menu da aplicação | Implementado em `main.go` com as três consultas, inserção, remoção e exibição |
| Busca de acessos, inserção, remoção lógica e manutenção por reconstrução no pacote `core` | Implementadas |
| Consultas de produto, acesso e relacionamento | Implementadas na aplicação principal e em `core` |

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
│   ├── storage.go          # Abertura, criação, leitura binária, tamanho e Seek
│   ├── index.go            # Buscas indexadas e consultas de relacionamento
│   ├── direct_search.go    # Buscas binárias diretas nos arquivos .dat
│   ├── queries.go          # Inserção, remoção lógica e reativação
│   ├── display.go          # Exibição parcial ou total dos registros
│   └── rebuild.go          # Reconstrução de dados, índices e extensões
├── scripts/
│   └── formatarDados.py    # Limpeza e conversão do CSV
├── .gitignore
├── go.mod
├── main.go
└── README.md
```

O CSV original permanece local e não é versionado. Para demonstrar os resultados da preparação, foram enviados ao Git amostras dos arquivos de dados e dos índices gerados:

```text
csvs/
└── 2019-Oct.csv            # Entrada, fornecida pelo usuário
dados/                     # Pasta com os arquivos gerados e suas amostras versionadas
├── produtos.dat           # Arquivo de dados de produtos
├── acessos.dat            # Arquivo de dados de acessos
├── indice_prod.idx        # Índice parcial de produtos
└── indice_acess.idx        # Índice parcial de acessos
```

Os arquivos enviados ao Git são amostras representativas, utilizadas para documentar e demonstrar a estrutura dos arquivos binários e dos índices. O processamento completo pode ser reproduzido a partir do CSV original e do script de preparação.

## Pré-requisitos

- Python 3 com suporte a f-strings (3.6 ou superior). O script utiliza apenas `csv`, `os`, `struct` e `sys`, da biblioteca padrão; não há dependências para instalar nem arquivo `requirements.txt`.
- Uma instalação do Go capaz de atender à versão `1.27.2`, declarada no `go.mod`. Confira essa exigência ao preparar o ambiente; a indicação anterior de Go 1.20 não corresponde ao módulo atual.
- O arquivo `2019-Oct.csv`, do dataset de comportamento de comércio eletrônico do Kaggle citado no script.
- Espaço em disco para o CSV, os dois arquivos de dados e seus índices. O consumo de memória do processamento de produtos cresce com a quantidade de produtos distintos.

## Preparação e execução

### 1. Disponibilizar o CSV

Coloque manualmente o arquivo `2019-Oct.csv` na pasta `csvs/`.

O CSV deve estar em UTF-8, usar vírgula como separador e conter as seguintes colunas no cabeçalho, em qualquer ordem:

```csv
event_time,event_type,product_id,category_code,brand,price,user_id
2019-10-01 00:00:00 UTC,view,1001,electronics.smartphone,marca_exemplo,199.90,5001
```

Colunas adicionais são ignoradas. Para usar outros caminhos, altere `CAMINHO_CSV`, `CAMINHO_PRODUTOS_DAT` e `CAMINHO_ACESSOS_DAT` em `scripts/formatarDados.py`.

### 2. Gerar os arquivos binários

Execute o script `scripts/formatarDados.py`. Ele cria `dados/` automaticamente e percorre o CSV duas vezes: primeiro para gerar produtos e depois para gerar acessos. Os arquivos `.dat` de destino são sobrescritos. Ao final de cada etapa, exibe o tamanho do arquivo, a quantidade de registros e o resto da divisão pelo tamanho do registro.

### 3. Validar os dados

O validador percorre os dois arquivos e verifica:

- Tamanho total múltiplo de 106 bytes para produtos e de 70 bytes para acessos.
- Presença de pelo menos um registro em cada arquivo.
- IDs em ordem estritamente crescente, sem duplicatas.
- Delimitador `\n` em todos os registros.
- Quantidade de registros lidos compatível com o tamanho do arquivo e término em EOF.

A leitura de acessos usa um buffer de 4 MiB. O programa informa o resultado de cada arquivo no terminal e indica quando alguma validação falha. Ele também verifica os offsets e as chaves dos índices em relação ao primeiro registro de cada bloco. A validação não verifica referências semânticas entre acessos e produtos.

### 4. Gerar os índices parciais

O programa lê os arquivos `.dat` e cria os índices que ainda não existem. Ele informa o resultado de cada geração e indica quando ocorre uma falha.

Após gerar novamente os dados, remova manualmente os índices antigos antes de gerar os novos. Índices existentes são ignorados, mesmo que estejam vazios ou desatualizados.

### Aplicação principal

A aplicação principal oferece consultas de produtos, acessos e relacionamento, inserção, remoção e exibição dos registros. Os índices devem existir e corresponder aos arquivos de dados antes da utilização.

As opções de pesquisa do menu são:

- consulta de produto por índice parcial;
- consulta de acesso por índice parcial;
- pesquisa binária direta de produto no arquivo de dados;
- pesquisa binária direta de acesso no arquivo de dados.

As quatro pesquisas exibem o tempo medido desde o início até o fim da operação de busca.

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

### `produtos.dat`: 106 bytes por registro

Formato Python: `<q32s48sdq?1s`. Estrutura Go: `core.ProdutoRecord`.

| Campo | Tipo Go | Offset | Bytes |
| --- | --- | ---: | ---: |
| `ProductID` | `int64` | 0 | 8 |
| `Brand` | `[32]byte` | 8 | 32 |
| `CategoryCode` | `[48]byte` | 40 | 48 |
| `Price` | `float64` | 88 | 8 |
| `Elo` | `int64` | 96 | 8 |
| `Excluido` | `bool` | 104 | 1 |
| `Delimiter` | `byte` (`0x0A`) | 105 | 1 |

### `acessos.dat`: 70 bytes por registro

Formato Python: `<qq16sq20sq?1s`. Estrutura Go: `core.AcessoRecord`.

| Campo | Tipo Go | Offset | Bytes |
| --- | --- | ---: | ---: |
| `EventID` | `int64` | 0 | 8 |
| `UserID` | `int64` | 8 | 8 |
| `EventType` | `[16]byte` | 16 | 16 |
| `ProductID` | `int64` | 32 | 8 |
| `EventTime` | `[20]byte` | 40 | 20 |
| `Elo` | `int64` | 60 | 8 |
| `Excluido` | `bool` | 68 | 1 |
| `Delimiter` | `byte` (`0x0A`) | 69 | 1 |

Tipos de evento esperados incluem `view`, `cart` e `purchase`, mas o script não restringe os valores a essa lista. Uma data como `2019-10-01 00:00:00` ocupa 19 caracteres e recebe um espaço para completar o campo.

O byte final `\n` faz parte do registro binário; os arquivos devem ser lidos pelo tamanho fixo dos registros, pois outros campos também podem conter o byte `0x0A`.

### Índices: 16 bytes por entrada

| Arquivo | Estrutura Go | Bytes 0–7 | Bytes 8–15 |
| --- | --- | --- | --- |
| `indice_prod.idx` | `core.ProdutoIdx` | `ProductID` (`int64`) | `Offset` (`int64`) |
| `indice_acess.idx` | `core.AcessoIdx` | `EventID` (`int64`) | `Offset` (`int64`) |

Os índices não possuem delimitador. Cada entrada referencia o primeiro registro de um bloco de **1.024 registros**, incluindo o primeiro bloco do arquivo. O valor `blockSize = 1024` em `cmd/indices/main.go` representa uma quantidade de registros, não de bytes.

Para a entrada de índice `k`, contando a partir de zero:

```text
offset em produtos.dat = k × 1024 × 106
offset em acessos.dat  = k × 1024 × 70
quantidade de entradas para N registros = teto(N / 1024)
tamanho do índice = quantidade de entradas × 16 bytes
```

Por exemplo, 1.025 registros geram duas entradas: uma para o registro de posição 0 e outra para o de posição 1.024. O offset de qualquer registro de posição `i` é `i × tamanho_do_registro`.

A consulta de produto localiza por busca binária a entrada de índice cuja chave seja a maior que não ultrapasse a chave procurada, posiciona `produtos.dat` com `Seek` e percorre até 1.024 registros do bloco. As entradas consultadas também são comparadas com seus registros de dados para detectar índices desatualizados ou inconsistentes.

### Área de extensão e reconstrução

Inserções que não encontram a chave no arquivo principal são gravadas na extensão do bloco correspondente. O campo `Elo` armazena o offset do próximo registro da cadeia dentro da extensão. A busca consulta o bloco principal e percorre essa cadeia quando necessário.

Cada extensão é verificada após uma inserção. Quando ela atinge 11 registros, o arquivo de dados e o índice daquele tipo são reconstruídos integralmente:

1. registros ativos do arquivo principal e das extensões são reunidos;
2. os registros são reorganizados pela chave;
3. registros com exclusão lógica são ignorados;
4. um novo arquivo temporário de dados é gravado;
5. um novo índice parcial é gravado no mesmo processo, a cada 1.024 registros;
6. os arquivos temporários substituem os arquivos originais;
7. as extensões são removidas.

Durante a reconstrução, os elos dos registros ficam em `-1`, pois as extensões deixam de existir.

## Verificação

A validação dos arquivos gerados é feita pelo programa em `cmd/validador`. Ela verifica os arquivos de dados e a correspondência entre os índices e os primeiros registros de cada bloco.

As consultas previstas são:

- Obter a marca e o preço de um produto pelo ID.
- Obter a data/hora e o tipo de um evento pelo ID.
- Obter a marca do produto associado a um evento e a data desse evento.

As três consultas previstas estão disponíveis no menu:

- opção 1: marca, categoria e preço do produto pelo ID;
- opção 8: tipo e data/hora do evento pelo ID;
- opção 9: marca do produto associado ao evento e data/hora do evento.

A opção 9 realiza duas consultas indexadas: primeiro no índice de acessos e depois no índice de produtos.

## Hipóteses

- Produtos da mesma marca podem possuir preços diferentes.
- O tipo do evento estará relacionado a ações como `view`, `cart` ou `purchase`.
- O `ProductID` do acesso deve corresponder a um produto existente.
