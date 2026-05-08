package infakt_test

import (
	"context"
	"fmt"
	"log"

	infakt "github.com/przemekperon/infakt-go-sdk"
)

func ExampleNewClient() {
	client := infakt.NewClient("your-api-key")
	_ = client
	// Output:
}

func ExampleNewClient_withOptions() {
	client := infakt.NewClient("your-api-key",
		infakt.WithUserAgent("my-app/1.0"),
	)
	_ = client
	// Output:
}

func ExampleInvoiceService_List() {
	client := infakt.NewClient("your-api-key")

	invoices, meta, err := client.Invoices.List(context.Background(), &infakt.InvoiceListOptions{
		ListOptions: infakt.ListOptions{Limit: 10},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d invoices (total: %d)\n", len(invoices), meta.TotalCount)
	for _, inv := range invoices {
		fmt.Printf("  %s - %d\n", inv.Number, inv.GrossPrice)
	}
}

func ExampleClientEntityService_Create() {
	client := infakt.NewClient("your-api-key")

	entity, err := client.Clients.Create(context.Background(), &infakt.ClientEntity{
		CompanyName: "ACME Sp. z o.o.",
		NIP:         "1234567890",
		Street:      "ul. Testowa 1",
		City:        "Warszawa",
		PostalCode:  "00-001",
		Country:     "PL",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created client: %s (ID: %d)\n", entity.CompanyName, entity.ID)
}

func ExampleProductService_List() {
	client := infakt.NewClient("your-api-key")

	products, _, err := client.Products.List(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range products {
		fmt.Printf("%s - %d PLN\n", p.Name, p.UnitNetPrice)
	}
}

func ExampleString() {
	name := infakt.String("Test Product")
	fmt.Println(*name)
	// Output: Test Product
}

func ExampleInt() {
	price := infakt.Int(10000)
	fmt.Println(*price)
	// Output: 10000
}
