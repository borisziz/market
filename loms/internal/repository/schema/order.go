package schema

type Order struct {
	ID     int64  `db:"id"`
	Status string `db:"status"`
	User   int64  `db:"user_id"`
}

type OrderItem struct {
	Sku   uint32 `db:"sku"`
	Count uint16 `db:"count"`
}

type SoldedItem struct {
	Sku         uint32 `db:"sku"`
	Count       uint16 `db:"count"`
	WarehouseID int64  `db:"warehouse_id"`
}
