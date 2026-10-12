package main

import (
	"fmt"
)

// Ponto de entrada, menu interativo e chamadas principais.
func main() {
	fmt.Println("=== Sistema de Arquivo Sequencial-Indexado ===")
	fmt.Println("1. Inserir dados iniciais (Carga)")
	fmt.Println("2. Mostrar dados (Produtos / Acessos)")
	fmt.Println("3. Pesquisa binária direta no arquivo de dados")
	fmt.Println("4. Consultar dados a partir do índice parcial")
	fmt.Println("5. Inserir registro")
	fmt.Println("6. Remover registro")
	fmt.Println("7. Consultas analíticas")
	fmt.Println("8. Reconstruir índices parciais")
	fmt.Println("0. Sair")
}
