package database

import (
	"context"
	"github.com/KalashnikovProjects/RamGenerator/internal/entities"
	_ "github.com/lib/pq" // Импортируем драйвер PostgreSQL
)

func GetUserContext(ctx context.Context, db SQLQueryExec, id int) (entities.User, error) {
	row := db.QueryRowContext(ctx, "SELECT id, username, password_hash, last_ram_generated, avatar_ram_id, avatar_box FROM users WHERE id=$1", id)
	user := entities.User{}
	err := row.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.LastRamGenerated, &user.AvatarRamId, &user.AvatarBox)
	if err != nil {
		return entities.User{}, err
	}
	return user, nil
}

func GetUserByUsernameContext(ctx context.Context, db SQLQueryExec, username string) (entities.User, error) {
	row := db.QueryRowContext(ctx, "SELECT id, username, password_hash, last_ram_generated, avatar_ram_id, avatar_box FROM users WHERE username=$1", username)
	user := entities.User{}
	err := row.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.LastRamGenerated, &user.AvatarRamId, &user.AvatarBox)
	if err != nil {
		return entities.User{}, err
	}
	return user, nil
}

// CreateUserContext создаёт пользователя и возвращает id
func CreateUserContext(ctx context.Context, db SQLQueryExec, user entities.User) (int, error) {
	var id int
	err := db.QueryRowContext(ctx, "INSERT INTO users (username, password_hash, last_ram_generated, avatar_ram_id, avatar_box) VALUES ($1, $2, $3, $4, $5)",
		user.Username, user.PasswordHash, user.LastRamGenerated, user.AvatarRamId, user.AvatarBox).Scan(id)
	return id, err
}

// PatchUserContext изменяет обновляет поля пользователя, nil поля в передаваемом user игнорируются
func PatchUserContext(ctx context.Context, db SQLQueryExec, user entities.User) {

}
