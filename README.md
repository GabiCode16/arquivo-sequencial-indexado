# Trabalho 1: Arquivos de Dados com Índices Parciais (AED 2)

Projeto desenvolvido para a disciplina de **Algoritmos e Estruturas de Dados II**, implementando uma organização de arquivos **sequencial-indexada** com índices parciais e registros de tamanho fixo em memória secundária.

---

## Estrutura do Projeto

```text
├── main.go               # Ponto de entrada, menu interativo e chamadas principais
├── core/
│   ├── model.go          # Definição das structs (produtos e acessos) e regras de tamanho fixo
│   ├── storage.go        # Funções de manipulação de arquivos binários, Seek, inserção e remoção
│   ├── index.go          # Gerenciamento dos índices parciais, busca binária e reconstrução
│   └── queries.go        # Consultas analíticas e respostas às perguntas do trabalho
├── dados/
│   ├── produtos.dat      # Arquivo binário de produtos de tamanho fixo
│   ├── acessos.dat       # Arquivo binário de acessos de tamanho fixo
│   ├── indice_prod.idx   # Arquivo de índice parcial de produtos
│   └── indice_acess.idx  # Arquivo de índice parcial de acessos
├── csvs/                 # Onde você coloca o dataset original do Kaggle
├── go.mod
└── README.md             # Documentação exigida para entrega no AVA
```

---

## Contexto do Trabalho e Dataset

- **Fonte dos Dados:** [eCommerce behavior data from multi-category store (Kaggle)](https://www.kaggle.com/datasets/mkechinov/ecommerce-behavior-data-from-multi-category-store)
- **Domínio:** Comportamento e eventos de usuários em loja virtual de departamentos (visualização, carrinho, compra, produtos, categorias e preços).

---

## Arquivos de Dados e Estruturas

### 1. `produtos.dat` (Arquivo de Produtos)
- **Organização:** Sequencial ordenado pelo campo chave.
- **Tamanho Fixo:** Campos de texto com preenchimento de espaços (*padding*) e cada registro finalizado com `\n`.
- **Campos:**
  - `ID`: Chave primária única (não repetida, numérico).
  - `Categoria`: Categoria hierárquica do produto.
  - `Marca`: Nome da marca/fabricante.
  - `Preco`: Valor do produto em ponto flutuante.
  - `Ativo`: Flag de controle para deleção lógica.

### 2. `acessos.dat` (Arquivo de Acessos ao E-commerce)
- **Organização:** Sequencial ordenado pelo campo chave.
- **Tamanho Fixo:** Padronizado em bytes por registro finalizado com `\n`.
- **Campos:**
  - `ID`: Chave sequencial única gerada por incremento.
  - `UsuarioID`: Identificador do usuário que realizou o evento.
  - `ProdutoID`: Identificador do produto acessado.
  - `TipoEvento`: Tipo de interação (`view`, `cart`, `purchase`).
  - `DataHora`: Carimbo temporal do acesso.
  - `Ativo`: Flag de controle.

---

## Índices Parciais

- **`indice_prod.idx`**: Índice parcial ordenado mapeando blocos de registros do arquivo `produtos.dat`.
- **`indice_acess.idx`**: Índice parcial ordenado mapeando blocos de registros do arquivo `acessos.dat`.
- **Funcionamento:**
  - Guarda a chave de início de cada bloco e o offset correspondente no arquivo de dados.
  - Busca realizada em dois níveis: busca binária no arquivo de índice e posicionamento via `seek` no arquivo de dados para varredura do bloco.
  - Reconstrução acionada após modificações estruturais (inserções e remoções).

---

## Perguntas Analíticas (Consultas)

1. **Pergunta 1:** Qual a marca com maior volume de produtos cadastrados e sua média de preço?
   - *Hipótese:* Marcas consolidadas concentram a maior fatia do catálogo com faixas de preço diversificadas.
2. **Pergunta 2:** Quais foram as interações de um determinado usuário ou produto?
   - *Hipótese:* A proporção de eventos do tipo `view` é significativamente superior aos eventos de `purchase`.
3. **Pergunta 3:** Qual a taxa de conversão (compras / visualizações) por categoria de produto?

---

## Como Executar

```bash
# Executar o programa
go run .

# Ou compilar o binário
go build -o trabalho-aed2.exe
./trabalho-aed2.exe
```

