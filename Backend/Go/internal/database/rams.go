package database

import (
	"context"
	"github.com/KalashnikovProjects/RamGenerator/internal/entities"
	_ "github.com/lib/pq"
)

func GetRamContext(ctx context.Context, db SQLQueryExec, id int) (entities.Ram, error) {
	query := `SELECT id, description, image_url, user_id FROM rams 
                                           WHERE id=$1`
	row := db.QueryRowContext(ctx, query, id)
	ram := entities.Ram{}
	err := row.Scan(&ram.Id, &ram.Description, &ram.ImageUrl, &ram.UserId)
	if err != nil {
		return entities.Ram{}, err
	}
	return ram, nil
}

func GetRamsByUsernameContext(ctx context.Context, db SQLQueryExec, username string) ([]entities.Ram, error) {
	query := `SELECT r.id, r.description, r.image_url, r.user_id FROM rams AS r 
    								 LEFT JOIN users AS u ON u.id=r.user_id 
                                                   WHERE u.username=$1`
	rows, err := db.QueryContext(ctx, query, username)
	if err != nil {
		return nil, err
	}
	var res []entities.Ram
	defer rows.Close()
	for rows.Next() {
		ram := entities.Ram{}
		err := rows.Scan(&ram.Id, &ram.Description, &ram.ImageUrl, &ram.UserId)
		if err != nil {
			return []entities.Ram{}, err
		}
		res = append(res, ram)
	}
	return res, nil
}

func GetRamsByUserIdContext(ctx context.Context, db SQLQueryExec, userId int) ([]entities.Ram, error) {
	query := `SELECT id, description, image_url, user_id FROM rams 
                                           WHERE user_id=$1`
	rows, err := db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	var res []entities.Ram
	for rows.Next() {
		ram := entities.Ram{}
		err := rows.Scan(&ram.Id, &ram.Description, &ram.ImageUrl, &ram.UserId)
		if err != nil {
			return []entities.Ram{}, err
		}
		res = append(res, ram)
	}
	return res, nil
}

// CreateRamContext создаёт барана и возвращает id
func CreateRamContext(ctx context.Context, db SQLQueryExec, ram entities.Ram) (int, error) {
	var id int
	err := db.QueryRowContext(ctx, "INSERT INTO rams (description, image_url, user_id) VALUES ($1, $2, $3) RETURNING id",
		ram.Description, ram.ImageUrl, ram.UserId).Scan(&id)
	return id, err
}

// UpdateRamContext изменяет поля пользователя, nil поля в передаваемом user игнорируются
func UpdateRamContext(ctx context.Context, db SQLQueryExec, id int, ram entities.Ram) error {
	fields := map[string]any{
		"description": ram.Description,
		"image_url":   ram.ImageUrl,
		"user_id":     ram.UserId,
	}
	query, args := GenerateQueryAndArgsForUpdate("rams", fields, "id=$1", id)
	_, err := db.ExecContext(ctx, query, args...)
	return err
}

func DeleteRamContext(ctx context.Context, db SQLQueryExec, id int) error {
	_, err := db.ExecContext(ctx, "DELETE FROM rams WHERE id=$1", id)
	return err
}
