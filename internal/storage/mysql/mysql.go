package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"hs-conclusion/model"
	"log"
	"strings"

	_ "github.com/microsoft/go-mssqldb"
)

type Storage struct {
	db *sql.DB
}

func New() (*Storage, error) {
	const op = "storage.mssql.NewStorage" // Имя текущей функции для логов и ошибок

	server := "wms.grass.local"
	user := "user_user"
	password := "lolololollo"
	database := "wms"

	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s", server, user, password, database)
	db, err := sql.Open("sqlserver", connString)
	if err != nil {

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	ctx := context.Background()
	err = db.PingContext(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("Connected!")

	return &Storage{db: db}, nil

}

func (s *Storage) PingDB() {
	const op = "storage.sqlite.SaveURL"

	pingErr := s.db.Ping()
	if pingErr != nil {

		fmt.Errorf("%s: %w", op, pingErr)
	}
	fmt.Println("Connected PingDB!")

}

func (s *Storage) GetPallets(barcode string) ([]model.Pallet, error) {
	connString := fmt.Sprintf(`exec sp_ReturnTreeTS '%s'`, strings.Replace(barcode, `\`, "", -1))
	rows, err := s.db.Query(connString)
	if err != nil {
		return nil, fmt.Errorf("getPalletByBarcode %q: %v", barcode, err)
	}
	defer rows.Close()
	// Loop through rows, using Scan to assign column data to struct fields.
	var pallets = []model.Pallet{}
	pallets, err = mapToPallets(rows)

	if err != nil {
		return nil, fmt.Errorf("getPalletByBarcode %q: %v", barcode, err)
	}
	return pallets, nil
}

func mapToPallets(rows *sql.Rows) ([]model.Pallet, error) {
	var palletsMap = make(map[string]*model.Pallet)
	var pallet *model.Pallet
	var countBox, countItem int

	for rows.Next() {
		var palletBarcode, boxBarcode, itemBarcode string
		var box model.Box
		var listItems []model.Item

		if err := rows.Scan(&palletBarcode, &boxBarcode, &itemBarcode, &countBox, &countItem); err != nil {
			return nil, err
		}

		if pallet == nil || pallet.Barcode != palletBarcode {
			pallet = &model.Pallet{
				Barcode:        palletBarcode,
				CountItemInBox: countBox,
			}
			palletsMap[palletBarcode] = pallet
		}

		for i := 0; i < countItem; i++ {
			rows.Scan(&palletBarcode, &boxBarcode, &itemBarcode, &countBox, &countItem)
			listItems = append(listItems, model.Item{Barcode: itemBarcode})
			rows.Next()

		}
		box = model.Box{
			Barcode: boxBarcode,
			Items:   listItems,
		}
		pallet.Items = append(pallet.Items, box)
	}

	var pallets []model.Pallet
	for _, p := range palletsMap {
		pallets = append(pallets, *p)
	}

	return pallets, nil
}

func (s *Storage) InsertDataIntoDB(pallet model.Pallet) error {
	// Begin a database transaction
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	connString := fmt.Sprintf(`INSERT INTO r_QrCodeUpdate (cCodeShort, idQrParent) VALUES (%s, 0)`, pallet.Barcode)
	// Insert Pallet data
	_, err =
		tx.Exec(connString)
		// tx.Exec("INSERT INTO r_QrCodeUpdate (cCodeShort, idQrParent) VALUES (?,0)", pallet.Barcode)
	if err != nil {
		tx.Rollback()
		return err
	}
	// Get Pallet ID in table
	// palletID, err := result.LastInsertId()
	// if err != nil {
	// 	tx.Rollback()
	// 	return err
	// }

	// Insert Box data
	for _, box := range pallet.Items {
		_, err := tx.Exec("INSERT INTO r_QrCodeUpdate (cCodeShort, idQrParent) VALUES (@boxBarcode, @idParent)", sql.Named("boxBarcode", box.Barcode), sql.Named("idParent", pallet.Barcode))

		if err != nil {
			tx.Rollback()
			return err
		}
		// Get Box ID in table
		// boxID, err := result.LastInsertId()
		// if err != nil {
		// 	tx.Rollback()
		// 	return err
		// }

		// Insert Item data
		for _, item := range box.Items {
			_, err := tx.Exec("INSERT INTO r_QrCodeUpdate (cCodeShort, idQrParent) VALUES (@itemBarcode, @idParent)", sql.Named("itemBarcode", item.Barcode), sql.Named("idParent", box.Barcode))
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	// Commit the transaction if all inserts were successful
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
