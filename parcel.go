package main

import (
	"database/sql"
	"log/slog"
)

type ParcelStore struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewParcelStore(db *sql.DB, logger *slog.Logger) ParcelStore {
	return ParcelStore{
		db:     db,
		logger: logger,
	}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	row := s.db.QueryRow(
		`
		insert into parcel (client, status, address, created_at)
		values (?, ?, ?, ?) 
		returning *
		`, p.Client, p.Status, p.Address, p.CreatedAt)

	parcel := Parcel{}
	err := row.Scan(&parcel.Number, &parcel.Client, &parcel.Status, &parcel.Address, &parcel.CreatedAt)
	if err != nil {
		s.logger.Error("failed to add parcel", "error", err)
		return 0, err
	}

	return parcel.Number, nil
}

func (s ParcelStore) Get(number int) (*Parcel, error) {
	row := s.db.QueryRow(
		`
		select 
    	number, client, status, address, created_at 
		from parcel 
		where number = $1
		`, number)

	parcel := Parcel{}
	err := row.Scan(&parcel.Number, &parcel.Client, &parcel.Status, &parcel.Address, &parcel.CreatedAt)
	if err != nil {
		s.logger.Error("failed to get parcel", "error", err)
		return nil, err
	}

	return &parcel, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	rows, err := s.db.Query(
		`
		select 
		number, client, status, address, created_at 
		from parcel 
		where client = $1
		`, client)

	if err != nil {
		s.logger.Error("failed to get parcels by client", "error", err)
		return nil, err
	}

	var res []Parcel
	for rows.Next() {
		parcel := Parcel{}
		errScan := rows.Scan(&parcel.Number, &parcel.Client, &parcel.Status, &parcel.Address, &parcel.CreatedAt)
		if errScan != nil {
			s.logger.Error("failed to scan parcel", "error", errScan)
			return nil, errScan
		}
		res = append(res, parcel)
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	stmt, err := s.db.Prepare(
		`
		update parcel
		set status = $1
		where number = $2
		`)

	if err != nil {
		s.logger.Error("failed to prepare statement", "error", err)
		return err
	}

	_, err = stmt.Exec(status, number)
	if err != nil {
		s.logger.Error("failed to execute statement", "error", err)
		return err
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered
	stmt, err := s.db.Prepare(
		`
		update parcel
		set address = $1
		where number = $2 and status = 'registered'
		`)
	if err != nil {
		s.logger.Error("failed to prepare statement", "error", err)
		return err
	}

	_, err = stmt.Exec(address, number)
	if err != nil {
		s.logger.Error("failed to execute statement", "error", err)
		return err
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered
	stmt, err := s.db.Prepare(
		`
		delete from parcel
		where number = $1 and status = 'registered'
		`)
	if err != nil {
		s.logger.Error("failed to prepare statement", "error", err)
		return err
	}

	_, err = stmt.Exec(number)
	if err != nil {
		s.logger.Error("failed to execute statement", "error", err)
		return err
	}

	return nil
}
