package main

import (
	"os"
)

const (
	idxProdPath  = "dados/indice_prod.idx"
	idxAcessPath = "dados/indice_acess.idx"
)

func exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func main() {

	existsProdutos, existsAcessos :=
		exists(idxProdPath), exists(idxAcessPath)

	if !existsProdutos {
		// Handle the case where the products index doesn't exist

	}

	if !existsAcessos {
		// Handle the case where the access index doesn't exist
	}
}
