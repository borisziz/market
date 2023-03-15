package schema

type Stock struct {
	WarehouseID int64  `db:"warehouse_id"`
	Count       uint64 `db:"count"`
}
