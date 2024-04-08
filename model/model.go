package model

// pallet represents data about a record pallet.
type Pallet struct {
	Barcode        string `json:"barcode"`
	CountItemInBox int    `json:"countItemInBox"`
	Items          []Box  `json:"boxes"`
}
type Box struct {
	Barcode string `json:"barcode"`
	Items   []Item `json:"items"`
}
type Item struct {
	Barcode string `json:"barcode"`
}
