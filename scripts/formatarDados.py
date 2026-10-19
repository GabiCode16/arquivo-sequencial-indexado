import csv
import os
import struct
import sys

# Caminhos padrão
CAMINHO_CSV = os.path.join("csvs", "2019-Oct.csv")
CAMINHO_PRODUTOS_DAT = os.path.join("dados", "produtos.dat")
CAMINHO_ACESSOS_DAT = os.path.join("dados", "acessos.dat")

# Constantes para Produtos (conforme README.md)
TAM_BRAND = 32
TAM_CATEGORY = 48
FORMATO_PRODUTO = "<q32s48sd1s"
TAMANHO_REGISTRO_PROD = 97  # 8 + 32 + 48 + 8 + 1

# Constantes para Acessos (conforme README.md)
TAM_EVENT_TYPE = 16
TAM_EVENT_TIME = 20
FORMATO_ACESSO = "<qq16sq20s1s"
TAMANHO_REGISTRO_ACESSO = 61  # 8 + 8 + 16 + 8 + 20 + 1

def verificar_arquivo(caminho):
    """Verifica se o arquivo CSV de entrada existe."""
    if not os.path.exists(caminho):
        print(f"[ERRO] Arquivo não encontrado: {caminho}")
        sys.exit(1)
    
    tamanho_gb = os.path.getsize(caminho) / (1024 ** 3)
    print(f"[INFO] Arquivo localizado: {caminho} ({tamanho_gb:.2f} GB)")

def limpar_texto(valor, tamanho_max):
    """
    Trata valores nulos ou vazios e remove espaços excedentes,
    limitando ao tamanho máximo permitido.
    """
    if valor is None:
        return ""
    texto = str(valor).strip()
    return texto[:tamanho_max]

def limpar_produto(raw_id, raw_brand, raw_cat, raw_price):
    """
    Converte e valida os tipos de dados do produto.
    Retorna um dicionário com os campos do produto ou None se inválido.
    """
    try:
        product_id = int(raw_id)
    except (ValueError, TypeError):
        return None

    try:
        price = float(raw_price) if raw_price else 0.0
    except (ValueError, TypeError):
        price = 0.0

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
    Lê o CSV em streaming, realiza a limpeza dos campos e
    deduplica os registros pelo campo chave 'product_id'.
    """
    verificar_arquivo(caminho_csv)
    print("[INFO] Iniciando extração, limpeza e deduplicação de produtos...")

    produtos = {}
    total_lidas = 0
    linhas_invalidas = 0

    with open(caminho_csv, mode="r", encoding="utf-8") as f:
        leitor = csv.reader(f)
        cabecalho = next(leitor)

        indices = {
            "product_id": cabecalho.index("product_id"),
            "category_code": cabecalho.index("category_code"),
            "brand": cabecalho.index("brand"),
            "price": cabecalho.index("price")
        }

        for linha in leitor:
            total_lidas += 1

            prod = limpar_produto(
                raw_id=linha[indices["product_id"]],
                raw_brand=linha[indices["brand"]],
                raw_cat=linha[indices["category_code"]],
                raw_price=linha[indices["price"]]
            )

            if prod is None:
                linhas_invalidas += 1
                continue

            pid = prod["product_id"]

            if pid not in produtos:
                produtos[pid] = prod
            else:
                if not produtos[pid]["brand"] and prod["brand"]:
                    produtos[pid]["brand"] = prod["brand"]
                if not produtos[pid]["category_code"] and prod["category_code"]:
                    produtos[pid]["category_code"] = prod["category_code"]

            if total_lidas % 1_000_000 == 0:
                print(f"[PROGRESSO] {total_lidas:,} linhas processadas | {len(produtos):,} produtos únicos...")

            if limite_linhas and total_lidas >= limite_linhas:
                print(f"[INFO] Limite de {limite_linhas:,} linhas atingido.")
                break

    print(f"\n[INFO] Leitura concluída: {total_lidas:,} linhas lidas, {len(produtos):,} produtos únicos.")
    return produtos

def gerar_produtos_dat(caminho_csv, caminho_dat, limite_linhas=None):
    """
    Extrai produtos únicos, ordena por product_id e grava no arquivo binário de 97 bytes.
    """
    produtos = extrair_produtos_unicos(caminho_csv, limite_linhas)

    print(f"[INFO] Ordenando {len(produtos):,} produtos por product_id...")
    produtos_ordenados = sorted(produtos.values(), key=lambda p: p["product_id"])

    os.makedirs(os.path.dirname(caminho_dat), exist_ok=True)
    print(f"[INFO] Gravando arquivo binário: {caminho_dat} ...")

    with open(caminho_dat, "wb") as f_out:
        for p in produtos_ordenados:
            brand_bytes = p["brand"].ljust(TAM_BRAND)[:TAM_BRAND].encode("ascii", "ignore")
            cat_bytes = p["category_code"].ljust(TAM_CATEGORY)[:TAM_CATEGORY].encode("ascii", "ignore")

            registro_binario = struct.pack(
                FORMATO_PRODUTO,
                p["product_id"],
                brand_bytes,
                cat_bytes,
                p["price"],
                b"\n"
            )
            f_out.write(registro_binario)

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
    print(f"Integridade (resto % 97): {resto} (0 = Perfeito)")
    print("=" * 55)

    with open(caminho_dat, "rb") as f_in:
        primeiro_buf = f_in.read(TAMANHO_REGISTRO_PROD)
        pid_pri, brand_pri, cat_pri, preco_pri, _ = struct.unpack(FORMATO_PRODUTO, primeiro_buf)
        print(f"\nPrimeiro registro (Índice 0):")
        print(f"  ID: {pid_pri} | Marca: '{brand_pri.decode('ascii').strip()}' | Preço: R$ {preco_pri:.2f}")

        f_in.seek((total_registros - 1) * TAMANHO_REGISTRO_PROD)
        ultimo_buf = f_in.read(TAMANHO_REGISTRO_PROD)
        pid_ult, brand_ult, cat_ult, preco_ult, _ = struct.unpack(FORMATO_PRODUTO, ultimo_buf)
        print(f"\nÚltimo registro (Índice {total_registros - 1:,}):")
        print(f"  ID: {pid_ult} | Marca: '{brand_ult.decode('ascii').strip()}' | Preço: R$ {preco_ult:.2f}")

def gerar_acessos_dat(caminho_csv, caminho_dat, limite_linhas=None, tamanho_buffer=100_000):
    """
    Gera o arquivo binário acessos.dat em streaming com registros fixos de 61 bytes.
    Chave primária (event_id) gerada sequencialmente garantindo ordenação natural.
    """
    verificar_arquivo(caminho_csv)
    os.makedirs(os.path.dirname(caminho_dat), exist_ok=True)
    print(f"[INFO] Iniciando geração de acessos: {caminho_dat} ...")

    total_processadas = 0
    linhas_invalidas = 0
    event_id_seq = 0
    buffer_bytes = []

    with open(caminho_csv, mode="r", encoding="utf-8") as f_in, open(caminho_dat, mode="wb") as f_out:
        leitor = csv.reader(f_in)
        cabecalho = next(leitor)

        idx_time = cabecalho.index("event_time")
        idx_type = cabecalho.index("event_type")
        idx_prod = cabecalho.index("product_id")
        idx_user = cabecalho.index("user_id")

        for linha in leitor:
            total_processadas += 1

            # Conversão e validação
            try:
                user_id = int(linha[idx_user])
                product_id = int(linha[idx_prod])
            except (ValueError, TypeError, IndexError):
                linhas_invalidas += 1
                continue

            event_id_seq += 1

            # Formatação de texto com padding
            tipo_evento = linha[idx_type].strip().ljust(TAM_EVENT_TYPE)[:TAM_EVENT_TYPE].encode("ascii", "ignore")
            
            # Data/hora ISO: remove ' UTC' e aplica padding para 20 bytes ("YYYY-MM-DD HH:MM:SS ")
            data_hora_limpa = linha[idx_time].replace(" UTC", "").strip()
            data_hora = data_hora_limpa.ljust(TAM_EVENT_TIME)[:TAM_EVENT_TIME].encode("ascii", "ignore")

            # Empacotamento de 61 bytes Little-Endian terminado com '\n'
            registro_binario = struct.pack(
                FORMATO_ACESSO,
                event_id_seq,
                user_id,
                tipo_evento,
                product_id,
                data_hora,
                b"\n"
            )
            buffer_bytes.append(registro_binario)

            # Gravação em lotes (buffer) para alto desempenho de I/O
            if len(buffer_bytes) >= tamanho_buffer:
                f_out.write(b"".join(buffer_bytes))
                buffer_bytes.clear()

            if total_processadas % 1_000_000 == 0:
                print(f"[PROGRESSO] {total_processadas:,} linhas processadas | {event_id_seq:,} acessos gravados...")

            if limite_linhas and total_processadas >= limite_linhas:
                print(f"[INFO] Limite de {limite_linhas:,} linhas atingido.")
                break

        # Grava os registros restantes no buffer
        if buffer_bytes:
            f_out.write(b"".join(buffer_bytes))
            buffer_bytes.clear()

    # Validação da integridade do arquivo gerado
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
    print(f"Integridade (resto % 61): {resto} (0 = Perfeito)")
    print("=" * 55)

    with open(caminho_dat, "rb") as f_in:
        # Primeiro registro
        primeiro_buf = f_in.read(TAMANHO_REGISTRO_ACESSO)
        eid_p, uid_p, tipo_p, pid_p, data_p, _ = struct.unpack(FORMATO_ACESSO, primeiro_buf)
        print(f"\nPrimeiro registro (Índice 0):")
        print(f"  EventID: {eid_p} | UserID: {uid_p} | Tipo: '{tipo_p.decode('ascii').strip()}' | ProdID: {pid_p} | Data: '{data_p.decode('ascii').strip()}'")

        # Último registro
        f_in.seek((total_registros - 1) * TAMANHO_REGISTRO_ACESSO)
        ultimo_buf = f_in.read(TAMANHO_REGISTRO_ACESSO)
        eid_u, uid_u, tipo_u, pid_u, data_u, _ = struct.unpack(FORMATO_ACESSO, ultimo_buf)
        print(f"\nÚltimo registro (Índice {total_registros - 1:,}):")
        print(f"  EventID: {eid_u} | UserID: {uid_u} | Tipo: '{tipo_u.decode('ascii').strip()}' | ProdID: {pid_u} | Data: '{data_u.decode('ascii').strip()}'")

if __name__ == "__main__":
    # Executa a geração de acessos.dat a partir do CSV
    gerar_acessos_dat(CAMINHO_CSV, CAMINHO_ACESSOS_DAT, limite_linhas=None)
