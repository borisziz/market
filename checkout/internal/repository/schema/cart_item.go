package schema

type CartItem struct {
	Sku   uint32 `db:"sku"`
	Count uint16 `db:"count"`
}
