package repository

import (
	"database/sql"

	"github.com/nagohak/chat-app/models"
)

type Room struct {
	Id      string
	Name    string
	Private bool
}

func (room *Room) GetId() string {
	return room.Id
}

func (room *Room) GetName() string {
	return room.Name
}

func (room *Room) GetPrivate() bool {
	return room.Private
}

type roomRepository struct {
	db *sql.DB
}

func NewRoomRepository(db *sql.DB) models.RoomRepository {
	return &roomRepository{db: db}
}

func (repo *roomRepository) AddRoom(room models.Room) error {
	stmt, err := repo.db.Prepare("INSERT INTO rooms(id, name, private) values (?,?,?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(room.GetId(), room.GetName(), room.GetPrivate())
	if err != nil {
		return err
	}

	return nil
}

func (repo *roomRepository) FindRoomByName(name string) (models.Room, error) {
	row := repo.db.QueryRow("SELECT id, name, private FROM rooms WHERE name = ? LIMIT 1", name)

	var room Room

	if err := row.Scan(&room.Id, &room.Name, &room.Private); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &room, nil
}
