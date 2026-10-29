"""
================================================================================
SCRIPT DE PRÉ-PROCESSAMENTO E FORMATAÇÃO DE DADOS
Disciplina: Algoritmos e Estruturas de Dados II
Tema: Organização de Arquivos Sequencial-Indexada

Objetivo:
    Transformar o dataset bruto do Kaggle (2019-Oct.csv, ~5.28 GB, 42M+ linhas)
    em dois arquivos binários de registros com TAMANHO FIXO (produtos.dat e acessos.dat),
    estritamente ordenados por suas respectivas chaves primárias, compatíveis com
    a leitura direta via Seek e binary.Read da linguagem Go.

Arquivos gerados:
    1. dados/produtos.dat : 106 bytes fixos por registro, ordenado por product_id
    2. dados/acessos.dat  : 70 bytes fixos por registro, ordenado por event_id
================================================================================
"""

import csv
import os
import struct
import sys

# ==============================================================================
# CONFIGURAÇÃO DE CAMINHOS DOS ARQUIVOS
# ==============================================================================
# Caminho para o CSV de origem (Kaggle eCommerce behavior data)
CAMINHO_CSV = os.path.join("csvs", "2019-Oct.csv")

# Caminhos de saída para os arquivos binários em memória secundária
CAMINHO_PRODUTOS_DAT = os.path.join("dados", "produtos.dat")
CAMINHO_ACESSOS_DAT = os.path.join("dados", "acessos.dat")

# ==============================================================================
# ESPECIFICAÇÃO TÉCNICA: PRODUTOS (produtos.dat)
# Total por registro: 106 bytes
#
# Estrutura em Go correspondente:
#   type ProdutoRecord struct {
#       ProductID    int64     // 8 bytes (Little-Endian)
#       Brand        [32]byte  // 32 bytes (ASCII com padding de espaços)
#       CategoryCode [48]byte  // 48 bytes (ASCII com padding de espaços)
#       Price        float64   // 8 bytes (IEEE 754 precisão dupla)
#       Elo          int64    // 8 bytes (offset na extensao, -1 = sem elo)
#       Excluido     bool     // 1 byte
#       Delimiter    byte      // 1 byte  ('\n' = 0x0A)
#   }
# ==============================================================================
TAM_BRAND = 32               # Tamanho fixo do campo marca em bytes
TAM_CATEGORY = 48            # Tamanho fixo da categoria em bytes
FORMATO_PRODUTO = "<q32s48sdq?1s"  # Little-Endian: id, marca, categoria, preco, elo, exclusao, delimitador
TAMANHO_REGISTRO_PROD = 106   # 8 + 32 + 48 + 8 + 8 + 1 + 1 = 106 bytes

# ==============================================================================
# ESPECIFICAÇÃO TÉCNICA: ACESSOS (acessos.dat)
# Total por registro: 70 bytes
#
# Estrutura em Go correspondente:
#   type AcessoRecord struct {
#       EventID   int64     // 8 bytes (Chave primária sequencial)
#       UserID    int64     // 8 bytes (ID do usuário)
#       EventType [16]byte  // 16 bytes (ASCII com padding de espaços)
#       ProductID int64     // 8 bytes (ID do produto associado)
#       EventTime [20]byte  // 20 bytes (Data/hora ISO: YYYY-MM-DD HH:MM:SS )
#       Elo       int64     // 8 bytes (offset na extensao, -1 = sem elo)
#       Excluido  bool      // 1 byte
#       Delimiter byte      // 1 byte  ('\n' = 0x0A)
#   }
# ==============================================================================
TAM_EVENT_TYPE = 16          # Tamanho fixo do tipo de evento (view, cart, purchase)
TAM_EVENT_TIME = 20          # Tamanho fixo do timestamp
FORMATO_ACESSO = "<qq16sq20sq?1s"  # Little-Endian: id, usuario, tipo, produto, data, elo, exclusao, delimitador
TAMANHO_REGISTRO_ACESSO = 70 # 8 + 8 + 16 + 8 + 20 + 8 + 1 + 1 = 70 bytes


def verificar_arquivo(caminho):
    """
    Verifica a existência física do arquivo de entrada antes de iniciar o processamento.
    Exibe o tamanho em gigabytes (GB) para acompanhamento.
    """
    if not os.path.exists(caminho):
        print(f"[ERRO] Arquivo de origem não encontrado: {caminho}")
        print("[DICA] Certifique-se de que o arquivo CSV do Kaggle foi colocado na pasta 'csvs/'.")
        sys.exit(1)
    
    # Converte o tamanho em bytes para Gigabytes (GB)
    tamanho_gb = os.path.getsize(caminho) / (1024 ** 3)
    print(f"[INFO] Arquivo de origem localizado: {caminho} ({tamanho_gb:.2f} GB)")


def limpar_texto(valor, tamanho_max):
    """
    Padroniza strings para armazenamento em registros de tamanho fixo:
    1. Trata valores nulos (None) ou ausentes, transformando em string vazia "".
    2. Remove espaços em branco nas extremidades (strip).
    3. Trunca o texto caso exceda o tamanho máximo estipulado para a coluna.
    """
    if valor is None:
        return ""
    texto = str(valor).strip()
    return texto[:tamanho_max]


def limpar_produto(raw_id, raw_brand, raw_cat, raw_price):
    """
    Converte e valida os dados brutos de uma linha do CSV para a entidade Produto.
    
    Retorno:
        Dicionário com os tipos numéricos e textuais tratados, ou None se
        o product_id for inválido (garantindo que a chave primária seja consistente).
    """
    # Validação do ID numérico (chave primária)
    try:
        product_id = int(raw_id)
    except (ValueError, TypeError):
        return None

    # Validação e conversão do preço para ponto flutuante
    try:
        price = float(raw_price) if raw_price else 0.0
    except (ValueError, TypeError):
        price = 0.0

    # Limpeza e corte das strings com os tamanhos máximos permitidos
    brand = limpar_texto(raw_brand, TAM_BRAND)
    category_code = limpar_texto(raw_cat, TAM_CATEGORY)

    return {
        "product_id": product_id,
        "brand": brand,
        "category_code": category_code,
        "price": price
    }


def extrair_produtos_unicos(caminho_csv, limite_linhas=None):
    """
    Realiza a leitura do CSV (linha a linha) para identificar
    todos os produtos distintos presentes no dataset.
    
    Estratégia de Deduplicação:
        Como cada linha do CSV representa um evento de acesso, o mesmo produto
        aparece repetido milhares de vezes. Usamos um dicionário indexado pelo
        'product_id'. Se o produto já foi visto, mantemos o registro existente,
        mas enriquecemos campos de marca ou categoria caso linhas posteriores
        tragam esses dados preenchidos.
    """
    verificar_arquivo(caminho_csv)
    print("[INFO] Iniciando extração, limpeza e deduplicação de produtos...")

    produtos = {}       # Mapeia: product_id -> dicionário com dados do produto
    total_lidas = 0     # Contador de linhas lidas do CSV
    linhas_invalidas = 0  # Contador de linhas com ID corrompido ou ausente

    with open(caminho_csv, mode="r", encoding="utf-8") as f:
        leitor = csv.reader(f)
        
        # Leitura da primeira linha (cabeçalho) para mapeamento dinâmico das colunas
        cabecalho = next(leitor)

        # Mapeia dinamicamente os índices das colunas para evitar erros de posição
        indices = {
            "product_id": cabecalho.index("product_id"),
            "category_code": cabecalho.index("category_code"),
            "brand": cabecalho.index("brand"),
            "price": cabecalho.index("price")
        }

        # Iteração linha a linha sobre o arquivo
        for linha in leitor:
            total_lidas += 1

            # Converte e limpa os campos
            prod = limpar_produto(
                raw_id=linha[indices["product_id"]],
                raw_brand=linha[indices["brand"]],
                raw_cat=linha[indices["category_code"]],
                raw_price=linha[indices["price"]]
            )

            # Descarta linhas cujo ID seja inválido
            if prod is None:
                linhas_invalidas += 1
                continue

            pid = prod["product_id"]

            # Deduplicação com enriquecimento de dados
            if pid not in produtos:
                produtos[pid] = prod
            else:
                # Se o produto já existia com marca/categoria vazia e a nova linha tem valor, atualiza
                if not produtos[pid]["brand"] and prod["brand"]:
                    produtos[pid]["brand"] = prod["brand"]
                if not produtos[pid]["category_code"] and prod["category_code"]:
                    produtos[pid]["category_code"] = prod["category_code"]

            # Exibe feedback periódico de progresso a cada 1 milhão de linhas
            if total_lidas % 1_000_000 == 0:
                print(f"[PROGRESSO] {total_lidas:,} linhas processadas | {len(produtos):,} produtos únicos...")

            # Permite interromper a leitura para testes rápidos caso limite_linhas seja definido
            if limite_linhas and total_lidas >= limite_linhas:
                print(f"[INFO] Limite de teste atingido ({limite_linhas:,} linhas).")
                break

    print(f"\n[INFO] Leitura concluída: {total_lidas:,} linhas lidas, {len(produtos):,} produtos únicos encontrados.")
    return produtos


def gerar_produtos_dat(caminho_csv, caminho_dat, limite_linhas=None):
    """
    Gera o arquivo binário final 'dados/produtos.dat'.
    
    Etapas:
    1. Extração e deduplicação de produtos em streaming.
    2. Ordenação estrita em memória pela chave primária 'product_id' (crescente).
       Isso é um pré-requisito fundamental para a organização Sequencial-Indexada.
    3. Empacotamento binário com tamanho fixo (106 bytes por registro):
       - As strings são alinhadas à esquerda e preenchidas à direita com espaços ASCII (0x20).
       - O endianness é Little-Endian ('<'), padrão nativo da arquitetura x86_64 e do Go.
       - Cada registro termina com o caractere delimitador '\\n' (0x0A).
    4. Validação de integridade pós-gravação:
       - O tamanho do arquivo em bytes DEVE ser múltiplo exato de 106.
    """
    # 1. Obtenção dos produtos únicos
    produtos = extrair_produtos_unicos(caminho_csv, limite_linhas)

    # 2. Ordenação pela chave primária (exigência da organização sequencial)
    print(f"[INFO] Ordenando {len(produtos):,} produtos por product_id (ordem crescente)...")
    produtos_ordenados = sorted(produtos.values(), key=lambda p: p["product_id"])

    # 3. Gravação em disco em modo binário puro ("wb")
    os.makedirs(os.path.dirname(caminho_dat), exist_ok=True)
    print(f"[INFO] Gravando arquivo binário: {caminho_dat} ...")

    with open(caminho_dat, "wb") as f_out:
        for p in produtos_ordenados:
            # Aplica padding com espaços à direita (ljust) e converte para bytes ASCII
            brand_bytes = p["brand"].ljust(TAM_BRAND)[:TAM_BRAND].encode("ascii", "ignore")
            cat_bytes = p["category_code"].ljust(TAM_CATEGORY)[:TAM_CATEGORY].encode("ascii", "ignore")

            # Empacota no formato binário de 106 bytes fixos
            registro_binario = struct.pack(
                FORMATO_PRODUTO,
                p["product_id"],  # 8 bytes (int64)
                brand_bytes,      # 32 bytes
                cat_bytes,        # 48 bytes
                p["price"],       # 8 bytes (float64)
                -1,              # Elo: sem extensao
                False,           # Exclusao logica desativada
                b"\n"             # 1 byte
            )
            f_out.write(registro_binario)

    # 4. Validação da integridade estrutural do arquivo gerado
    tamanho_total = os.path.getsize(caminho_dat)
    total_registros = tamanho_total // TAMANHO_REGISTRO_PROD
    resto = tamanho_total % TAMANHO_REGISTRO_PROD

    print("\n" + "=" * 55)
    print("STATUS DO ARQUIVO BINÁRIO GERADO: produtos.dat")
    print("=" * 55)
    print(f"Caminho: {caminho_dat}")
    print(f"Tamanho total: {tamanho_total:,} bytes ({tamanho_total / (1024*1024):.2f} MB)")
    print(f"Total de registros gravados: {total_registros:,}")
    print(f"Tamanho por registro: {TAMANHO_REGISTRO_PROD} bytes fixos")
    print(f"Integridade (resto % 106): {resto} (0 = Perfeito)")
    print("=" * 55)


def gerar_acessos_dat(caminho_csv, caminho_dat, limite_linhas=None, tamanho_buffer=100_000):
    """
    Gera o arquivo binário final 'dados/acessos.dat'.
    
    Características e Desafios:
    1. Volume Massivo:
       O CSV contém mais de 42 milhões de eventos de navegação (~2.4 GB de binário).
    2. Bufferização em Lote (tamanho_buffer = 100.000):
       Escrever registro a registro no disco geraria 42 milhões de chamadas ao sistema
       operacional (syscalls de I/O), tornando o processo extremamente lento.
       Usamos uma lista na memória que acumula 100.000 registros binários e grava
       tudo de uma só vez com um único write(b''.join(...)), multiplicando a velocidade por 10x.
    3. Chave Primária Sequencial (event_id):
       Como o CSV original não possui um identificador numérico único simples para cada
       evento, geramos uma chave incremental sequencial (1, 2, 3... N), garantindo
       que o arquivo de dados já seja gravado perfeitamente ordenado pela chave primária.
    4. Formatação de 70 bytes fixos:
       - user_id e product_id convertidos para int64 (8 bytes cada).
       - event_type com padding de espaços até 16 bytes.
       - event_time padronizado para 20 bytes (ex: '2019-10-01 00:00:00 ').
       - delimitador '\\n' (1 byte) ao final.
    """
    verificar_arquivo(caminho_csv)
    os.makedirs(os.path.dirname(caminho_dat), exist_ok=True)
    print(f"[INFO] Iniciando geração de acessos: {caminho_dat} ...")

    total_processadas = 0   # Contador total de linhas do CSV
    linhas_invalidas = 0    # Contador de linhas com dados corrompidos
    event_id_seq = 0        # Chave primária gerada incrementalmente
    buffer_bytes = []       # Buffer para acumular registros em memória antes da escrita

    with open(caminho_csv, mode="r", encoding="utf-8") as f_in, open(caminho_dat, mode="wb") as f_out:
        leitor = csv.reader(f_in)
        cabecalho = next(leitor)

        # Mapeamento dinâmico dos índices das colunas de acesso
        idx_time = cabecalho.index("event_time")
        idx_type = cabecalho.index("event_type")
        idx_prod = cabecalho.index("product_id")
        idx_user = cabecalho.index("user_id")

        for linha in leitor:
            total_processadas += 1

            # Validação e conversão das chaves estrangeiras
            try:
                user_id = int(linha[idx_user])
                product_id = int(linha[idx_prod])
            except (ValueError, TypeError, IndexError):
                linhas_invalidas += 1
                continue

            # Incrementa o identificador único sequencial (chave primária)
            event_id_seq += 1

            # Aplica padding de espaços à direita no tipo de evento (16 bytes)
            tipo_evento = linha[idx_type].strip().ljust(TAM_EVENT_TYPE)[:TAM_EVENT_TYPE].encode("ascii", "ignore")
            
            # Trata o campo de data/hora: remove ' UTC' e aplica padding para 20 bytes
            data_hora_limpa = linha[idx_time].replace(" UTC", "").strip()
            data_hora = data_hora_limpa.ljust(TAM_EVENT_TIME)[:TAM_EVENT_TIME].encode("ascii", "ignore")

            # Empacotamento de 70 bytes Little-Endian terminado com '\n'
            registro_binario = struct.pack(
                FORMATO_ACESSO,
                event_id_seq,  # 8 bytes (int64)
                user_id,       # 8 bytes (int64)
                tipo_evento,   # 16 bytes
                product_id,    # 8 bytes (int64)
                data_hora,     # 20 bytes
                -1,           # Elo: sem extensao
                False,        # Exclusao logica desativada
                b"\n"          # 1 byte
            )
            buffer_bytes.append(registro_binario)

            # Esvazia o buffer no disco quando atingir o limite estipulado
            if len(buffer_bytes) >= tamanho_buffer:
                f_out.write(b"".join(buffer_bytes))
                buffer_bytes.clear()

            # Feedback de progresso a cada 1 milhão de linhas
            if total_processadas % 1_000_000 == 0:
                print(f"[PROGRESSO] {total_processadas:,} linhas processadas | {event_id_seq:,} acessos gravados...")

            # Limite opcional para testes rápidos
            if limite_linhas and total_processadas >= limite_linhas:
                print(f"[INFO] Limite de teste atingido ({limite_linhas:,} linhas).")
                break

        # Grava os registros restantes que ficaram no buffer após o loop
        if buffer_bytes:
            f_out.write(b"".join(buffer_bytes))
            buffer_bytes.clear()

    # Validação da integridade estrutural do arquivo de acessos
    tamanho_total = os.path.getsize(caminho_dat)
    total_registros = tamanho_total // TAMANHO_REGISTRO_ACESSO
    resto = tamanho_total % TAMANHO_REGISTRO_ACESSO

    print("\n" + "=" * 55)
    print("STATUS DO ARQUIVO BINÁRIO GERADO: acessos.dat")
    print("=" * 55)
    print(f"Caminho: {caminho_dat}")
    print(f"Tamanho total: {tamanho_total:,} bytes ({tamanho_total / (1024*1024):.2f} MB)")
    print(f"Total de registros gravados: {total_registros:,}")
    print(f"Linhas descartadas (inválidas): {linhas_invalidas:,}")
    print(f"Tamanho por registro: {TAMANHO_REGISTRO_ACESSO} bytes fixos")
    print(f"Integridade (resto % 70): {resto} (0 = Perfeito)")
    print("=" * 55)


def gerar_todos():
    """
    Executa o pipeline completo: gera o arquivo de produtos e em seguida
    o arquivo de acessos a partir do CSV.
    """
    print("\n>>> ETAPA 1/2: GERANDO PRODUTOS.DAT <<<")
    gerar_produtos_dat(CAMINHO_CSV, CAMINHO_PRODUTOS_DAT, limite_linhas=None)

    print("\n>>> ETAPA 2/2: GERANDO ACESSOS.DAT <<<")
    gerar_acessos_dat(CAMINHO_CSV, CAMINHO_ACESSOS_DAT, limite_linhas=None)


# ==============================================================================
# PONTO DE ENTRADA PRINCIPAL DA LINHA DE COMANDO
# ==============================================================================
if __name__ == "__main__":
  gerar_todos()
