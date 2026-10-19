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
├── csvs/                 # Onde você coloca o dataset original do Kaggle
├── scripts/
│   └── formatarDados.py    # Script utilitário em Python para geração/processamento dos arquivos .dat
├── requirements.txt      # Dependências das bibliotecas Python para os scripts
├── go.mod                # Gerenciador de módulos Go
└── README.md             # Especificação e documentação da arquitetura
```

## Como Executar

### Pré-requisitos
- [Go](https://go.dev/) instalado (versão 1.20 ou superior).
- [Python 3](https://www.python.org/) (opcional, para execução dos scripts de preparação de dados):
  ```bash
  # Criar e ativar o ambiente virtual (se ainda não existir)
  python -m venv venv
  .\venv\Scripts\activate   # Windows
  # source venv/bin/activate # Linux/macOS

  # Instalar dependências
  pip install -r requirements.txt
  ```

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
#### Perguntas iniciais

--> Qual a marca e o preço do produto de ID X?

--> Qual o horário do evento e o tipo do evento de ID X?

--> Qual é a marca do produto associado ao evento de ID X, e qual a sua data?

##### Estruturas

================================================================================
ESPECIFICACAO TECNICA DE ARMAZENAMENTO: produtos.dat
================================================================================

1. VISÃO GERAL DO REGISTRO
--------------------------------------------------------------------------------
- Arquivo: produtos.dat
- Modo de abertura: Binario puro (wb em Python, os.OpenFile em Go)
- Tamanho total por registro: 97 bytes fixos
- Delimitador final: \n (ASCII 0x0A / byte 10)
- Endianness (Ordem dos bytes): Little-Endian
- Ordenacao: Obras/Registros gravados estritamente em ordem crescente pelo product_id.


2. ESTRUTURA DO REGISTRO EM GO (DECLARAÇÃO DE STRUCT)
--------------------------------------------------------------------------------
Em Go, a struct deve mapear exatamente os 97 bytes em memoria sem campos de 
tamanho variavel (como slice ou string direta). Usam-se arrays fixos de bytes.

type ProdutoRecord struct {
    ProductID    int64     //  8 bytes - Inteiro com sinal (Little-Endian)
    Brand        [32]byte  // 32 bytes - String UTF-8/ASCII preenchida com espacos
    CategoryCode [48]byte  // 48 bytes - String UTF-8/ASCII preenchida com espacos
    Price        float64   //  8 bytes - Ponto flutuante precisao dupla (IEEE 754)
    Delimiter    byte      //  1 byte  - Caractere '\n' (0x0A)
}

Tamanho Total da Struct: 8 + 32 + 48 + 8 + 1 = 97 bytes.


3. MAPEAMENTO DE TIPOS E PADRÃO DE GRAVAÇÃO EM PYTHON
--------------------------------------------------------------------------------
Para que o Go consiga ler via binary.Read(file, binary.LittleEndian, &struct),
o Python deve utilizar a biblioteca nativa 'struct' com a seguinte string de 
formatacao:

Format String do struct.pack: "<q32s48sd1s"

Detalhamento dos codigos de formatacao:
- '<' : Define endianness para Little-Endian (obrigatorio para Go)
- 'q' : int64 em Go  -> int em Python (signed long long, 8 bytes)
- '32s': [32]byte em Go -> bytes de tamanho 32 em Python
- '48s': [48]byte em Go -> bytes de tamanho 48 em Python
- 'd' : float64 em Go -> float em Python (double precision, 8 bytes)
- '1s': byte em Go    -> b'\n' em Python (char/byte unico, 1 byte)


4. REGRAS DE TRATAMENTO DE DADOS NO PYTHON ANTES DE EMPACOTAR
--------------------------------------------------------------------------------
a) Tratar Valores Nulos/Ausentes:
   - Se brand ou category_code forem None/NaN, converter para string vazia "".

b) Preenchimento de Strings (Padding):
   - As strings devem ser alinhadas a esquerda e preenchidas com espacos ate o 
     tamanho exato do campo.
   - Para brand:   texto.ljust(32)[:32].encode('ascii', 'ignore')
   - Para category: texto.ljust(48)[:48].encode('ascii', 'ignore')
   - O caractere de preenchimento deve ser o espaco ASCII (0x20).

c) Conversao Numerica:
   - product_id deve ser convertido para int.
   - price deve ser convertido para float.

d) Delimitador:
   - Deve ser explicitamente o byte b'\n'.


5. FLUXO COMPLETO DE LER NO GO E LIMPAR OS ESPAÇOS
--------------------------------------------------------------------------------
Quando o Go le o registro de 97 bytes com binary.Read(), os campos Brand e 
CategoryCode estarao populados com os caracteres originais mais os espacos de 
padding.

Para extrair a string limpa em Go:
- brandLimpa := strings.TrimSpace(string(prod.Brand[:]))
- categoryLimpa := strings.TrimSpace(string(prod.CategoryCode[:]))
================================================================================

================================================================================
ESPECIFICACAO TECNICA DE ARMAZENAMENTO: acessos.dat
================================================================================

1. VISÃO GERAL DO REGISTRO
--------------------------------------------------------------------------------
- Arquivo: acessos.dat
- Modo de abertura: Binário puro (wb em Python, os.OpenFile em Go)
- Tamanho total por registro: 61 bytes fixos
- Delimitador final: \n (ASCII 0x0A / byte 10)
- Endianness (Ordem dos bytes): Little-Endian
- Ordenação: Registros gravados estritamente em ordem crescente pelo event_id.


2. ESTRUTURA DO REGISTRO EM GO (DECLARAÇÃO DE STRUCT)
--------------------------------------------------------------------------------
Em Go, a struct deve mapear exatamente os 61 bytes em memória sem campos de 
tamanho variável. Usam-se arrays fixos de bytes para textos e datas.

type AcessoRecord struct {
    EventID   int64     //  8 bytes - Inteiro com sinal (Chave primária)
    UserID    int64     //  8 bytes - Inteiro com sinal (ID do usuário)
    EventType [16]byte  // 16 bytes - String ASCII com padding ("view", "cart", "purchase")
    ProductID int64     //  8 bytes - Inteiro com sinal (ID do produto associado)
    EventTime [20]byte  // 20 bytes - String ASCII com a data/hora ("YYYY-MM-DD HH:MM:SS")
    Delimiter byte      //  1 byte  - Caractere '\n' (0x0A)
}

Tamanho Total da Struct: 8 + 8 + 16 + 8 + 20 + 1 = 61 bytes.


3. MAPEAMENTO DE TIPOS E PADRÃO DE GRAVAÇÃO EM PYTHON
--------------------------------------------------------------------------------
Format String do struct.pack em Python: "<qq16sq20s1s"

Detalhamento dos códigos de formatação:
- '<'  : Endianness Little-Endian (padrão Go)
- 'q'  : event_id  (int64 em Go  -> int em Python, 8 bytes)
- 'q'  : user_id   (int64 em Go  -> int em Python, 8 bytes)
- '16s': event_type ([16]byte em Go -> bytes de tamanho 16 em Python)
- 'q'  : product_id (int64 em Go -> int em Python, 8 bytes)
- '20s': event_time ([20]byte em Go -> bytes de tamanho 20 em Python)
- '1s' : delimiter  (byte em Go   -> b'\n' em Python, 1 byte)


4. REGRAS DE TRATAMENTO DE DADOS NO PYTHON ANTES DE EMPACOTAR
--------------------------------------------------------------------------------
a) EventType (16 bytes):
   - Tipos de eventos conhecidos: "view", "cart", "purchase".
   - Aplicar padding com espaços: event_type.ljust(16)[:16].encode('ascii')

b) EventTime (20 bytes):
   - Tratar a string da data para garantir o formato fixo de 20 caracteres:
     Exemplo: "2019-10-01 00:00:00 UTC" ou "2019-10-01 00:00:00 "
   - Aplicar padding com espaços: event_time.ljust(20)[:20].encode('ascii')

c) Inteiros (event_id, user_id, product_id):
   - Garantir conversão para int nativo antes de passar para o struct.pack.

d) Delimitador:
   - Byte fixo b'\n'.


5. JUSTIFICATIVA DOS TAMANHOS DE STRINGS
--------------------------------------------------------------------------------
- EventType (16 bytes): O maior texto comum é "purchase" (8 caracteres). 
  16 bytes garantem folga total para qualquer novo tipo de ação.
- EventTime (20 bytes): Cobre com precisão o padrão ISO de data e hora sem 
  fuso estendido (ex: "2019-10-01 00:00:00" tem 19 caracteres + 1 de espaço).


6. CÁLCULO DE OFFSET PARA O SEEK
--------------------------------------------------------------------------------
Como cada registro possui rigorosamente 61 bytes:
Offset em Bytes = Índice do Registro * 61
================================================================================