package dbrepo

import (
	"context"
	"time"

	"github.com/bopepsi/bookings/internal/models"
)

func (this *postgresDBRepo) AllUsers() bool {
	return true
}

func (this *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var newID int

	query := `
			insert into reservations 
				(first_name, last_name, email, phone, start_date, end_date, room_id,
			created_at, updated_at) values
				($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id;
			`

	err := this.DB.QueryRowContext(ctx, query, res.FirstName, res.LastName, res.Email,
		res.Phone, res.StartDate, res.EndDate, res.RoomID,
		time.Now(), time.Now()).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (this *postgresDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `
			insert into room_restrictions
				(start_date, end_date, room_id, reservation_id, restriction_id, created_at, updated_at) 
			values
				($1, $2, $3, $4, $5, $6, $7)	
			`
	_, err := this.DB.ExecContext(ctx, query, r.StartDate, r.EndDate, r.RoomID,
		r.ReservationID, r.RestrictionID, time.Now(), time.Now())

	if err != nil {
		return err
	}

	return nil
}

func (this *postgresDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomId int) (bool, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
			select 
				count(*)
			from 
				room_restrictions
			where
				room_id = $3
				$1 < end_date and $2 > start_date;
			`

	var numRows int
	err := this.DB.QueryRowContext(ctx, query, start, end, roomId).Scan(&numRows)

	if err != nil {
		return false, err
	}

	if numRows == 0 {
		return true, nil
	}

	return false, nil
}

func (this *postgresDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
			select 
				r.id, r.name
			from
				rooms r
			where
				r.id not in 
				(select rr.room_id from room_restrictions rr where $1 < rr.end_date and $2 > rr.start_date)
			`

	var rooms []models.Room
	rows, err := this.DB.QueryContext(ctx, query, start, end)

	if err != nil {
		return rooms, err
	}

	for rows.Next() {
		var room models.Room
		err := rows.Scan(&room.ID, &room.RoomName)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}
