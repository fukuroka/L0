package order

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"L0/internal/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	client db.Client
	logger Logger
}

func NewOrderRepository(client *pgxpool.Pool, logger Logger) *OrderRepository {
	return &OrderRepository{client: client, logger: logger}
}

func (r *OrderRepository) Save(ctx context.Context, order Order) (err error) {
	tx, err := r.client.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	orderQuery := `
		INSERT INTO orders (
			order_uid, track_number, entry,
			locale, internal_signature, customer_id,
			delivery_service, shardkey, sm_id,
			date_created, oof_shard
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT DO NOTHING
	`
	_, err = tx.Exec(ctx, orderQuery,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.ShardKey,
		order.SmID,
		order.DateCreated,
		order.OofShard,
	)
	if err != nil {
		return err
	}

	deliveryQuery := `
        INSERT INTO deliveries (
            order_uid, name, phone, zip, city, address, region, email
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT DO NOTHING
    `
	_, err = tx.Exec(ctx, deliveryQuery,
		order.OrderUID,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email,
	)
	if err != nil {
		return err
	}

	paymentQuery := `
        INSERT INTO payments (
            order_uid, transaction, request_id, currency, provider,
            amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT DO NOTHING
    `
	_, err = tx.Exec(ctx, paymentQuery,
		order.OrderUID,
		order.Payment.Transaction,
		order.Payment.RequestID,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDt,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	)
	if err != nil {
		return err
	}

	if len(order.Products) > 0 {
		createTemp := `CREATE TEMP TABLE temp_products (
			chrt_id int,
			track_number text,
			price int,
			rid text,
			name text,
			sale int,
			size text,
			total_price int,
			nm_id int,
			brand text,
			status int
		) ON COMMIT DROP;`
		if _, err = tx.Exec(ctx, createTemp); err != nil {
			return err
		}

		cols := []string{"chrt_id", "track_number", "price", "rid", "name", "sale", "size", "total_price", "nm_id", "brand", "status"}
		rows := make([][]interface{}, 0, len(order.Products))
		for _, product := range order.Products {
			rows = append(rows, []interface{}{
				product.ChrtID,
				product.TrackNumber,
				product.Price,
				product.Rid,
				product.Name,
				product.Sale,
				product.Size,
				product.TotalPrice,
				product.NmID,
				product.Brand,
				product.Status,
			})
		}

		if _, err = tx.CopyFrom(ctx, pgx.Identifier{"temp_products"}, cols, pgx.CopyFromRows(rows)); err != nil {
			return err
		}

		upsert := `INSERT INTO products (chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
		SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM temp_products
		ON CONFLICT DO NOTHING;`
		if _, err = tx.Exec(ctx, upsert); err != nil {
			return err
		}
	}

	err = tx.Commit(ctx)
	return err
}

func (r *OrderRepository) GetById(ctx context.Context, orderId string) (Order, error) {
	query := `
		SELECT 
			o.order_uid, o.track_number, o.entry,
			o.locale, o.internal_signature, o.customer_id,
			o.delivery_service, o.shardkey, o.sm_id,
			o.date_created, o.oof_shard,
			to_jsonb(d.*) AS delivery,
			to_jsonb(p.*) AS payment,
			COALESCE(json_agg(pr.*) FILTER (WHERE pr.chrt_id IS NOT NULL), '[]') AS products
		FROM orders o
		LEFT JOIN deliveries d ON o.order_uid = d.order_uid
		LEFT JOIN payments p ON o.order_uid = p.order_uid
		LEFT JOIN products pr ON o.track_number = pr.track_number
		WHERE o.order_uid = $1
		GROUP BY o.order_uid, d.*, p.*;
	`

	row := r.client.QueryRow(ctx, query, orderId)

	var (
		orderUID, trackNumber, entry, locale, internalSignature, customerID,
		deliveryService, shardKey, oofShard sql.NullString
		smID                                    sql.NullInt64
		dateCreated                             time.Time
		deliveryJSON, paymentJSON, productsJSON []byte
	)

	if err := row.Scan(&orderUID, &trackNumber, &entry, &locale, &internalSignature, &customerID,
		&deliveryService, &shardKey, &smID, &dateCreated, &oofShard,
		&deliveryJSON, &paymentJSON, &productsJSON); err != nil {
		if err == sql.ErrNoRows {
			return Order{}, fmt.Errorf("order not found")
		}
		return Order{}, err
	}

	ord := Order{}
	ord.OrderUID = orderUID.String
	ord.TrackNumber = trackNumber.String
	ord.Entry = entry.String
	ord.Locale = locale.String
	ord.InternalSignature = internalSignature.String
	ord.CustomerID = customerID.String
	ord.DeliveryService = deliveryService.String
	ord.ShardKey = shardKey.String
	if smID.Valid {
		ord.SmID = int(smID.Int64)
	}
	ord.DateCreated = dateCreated
	ord.OofShard = oofShard.String

	if len(deliveryJSON) > 0 {
		var d Delivery
		if err := json.Unmarshal(deliveryJSON, &d); err == nil {
			ord.Delivery = d
		}
	}
	if len(paymentJSON) > 0 {
		var p Payment
		if err := json.Unmarshal(paymentJSON, &p); err == nil {
			ord.Payment = p
		}
	}
	if len(productsJSON) > 0 {
		var prods []Product
		if err := json.Unmarshal(productsJSON, &prods); err == nil {
			ord.Products = prods
		}
	}

	return ord, nil
}

func (r *OrderRepository) GetLimit(ctx context.Context, limit int) ([]Order, error) {
	query := `
		SELECT
			o.order_uid, o.track_number, o.entry,
			o.locale, o.internal_signature, o.customer_id,
			o.delivery_service, o.shardkey, o.sm_id,
			o.date_created, o.oof_shard,
			to_jsonb(d.*) AS delivery,
			to_jsonb(p.*) AS payment,
			COALESCE(json_agg(pr.*) FILTER (WHERE pr.chrt_id IS NOT NULL), '[]') AS products
		FROM orders o
		LEFT JOIN deliveries d ON o.order_uid = d.order_uid
		LEFT JOIN payments p ON o.order_uid = p.order_uid
		LEFT JOIN products pr ON o.track_number = pr.track_number
		GROUP BY o.order_uid, d.*, p.*
		ORDER BY o.date_created DESC
		LIMIT $1
	`

	rows, err := r.client.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Order, 0)
	for rows.Next() {
		var (
			orderUID, trackNumber, entry, locale, internalSignature, customerID,
			deliveryService, shardKey, oofShard sql.NullString
			smID                                    sql.NullInt64
			dateCreated                             time.Time
			deliveryJSON, paymentJSON, productsJSON []byte
		)

		if err := rows.Scan(&orderUID, &trackNumber, &entry, &locale, &internalSignature, &customerID,
			&deliveryService, &shardKey, &smID, &dateCreated, &oofShard,
			&deliveryJSON, &paymentJSON, &productsJSON); err != nil {
			return nil, err
		}

		ord := Order{}
		ord.OrderUID = orderUID.String
		ord.TrackNumber = trackNumber.String
		ord.Entry = entry.String
		ord.Locale = locale.String
		ord.InternalSignature = internalSignature.String
		ord.CustomerID = customerID.String
		ord.DeliveryService = deliveryService.String
		ord.ShardKey = shardKey.String
		if smID.Valid {
			ord.SmID = int(smID.Int64)
		}
		ord.DateCreated = dateCreated
		ord.OofShard = oofShard.String

		if len(deliveryJSON) > 0 {
			var d Delivery
			_ = json.Unmarshal(deliveryJSON, &d)
			ord.Delivery = d
		}
		if len(paymentJSON) > 0 {
			var p Payment
			_ = json.Unmarshal(paymentJSON, &p)
			ord.Payment = p
		}
		if len(productsJSON) > 0 {
			var prods []Product
			_ = json.Unmarshal(productsJSON, &prods)
			ord.Products = prods
		}

		result = append(result, ord)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
