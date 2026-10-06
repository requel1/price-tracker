package main

import (
	"context"
	"fmt"
	"log"

	"github.com/requel1/price-tracker/internal/parser"
)

func main() {
	p := parser.NewChitaiGorodParser()
	price, err := p.Parse(context.Background(), "https://www.chitai-gorod.ru/product/vyzit-v-kacestve-zeny-geroa-tom-3-3183836")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Price: %.2f %s\n", price.Price, price.Currency)
}
