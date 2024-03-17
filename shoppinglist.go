package main

type ShoppingList struct {
	ID    string             `json:"id"`
	Name  string             `json:"name"`
	Items []ShoppingListItem `json:"items,omitempty"`
}

type ShoppingListItem struct {
	ID             string `json:"id"`
	ItemID         string `json:"itemId"`
	Item           Item   `json:"item"`
	Quantity       int    `json:"quantity"`
	QuantityTarget *int   `json:"quantityTarget"`
}
