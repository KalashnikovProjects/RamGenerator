package database

import (
	"context"
	"database/sql"
	"github.com/KalashnikovProjects/RamGenerator/internal/entities"
	_ "github.com/lib/pq" // Импортируем драйвер PostgreSQL
)

func GetUserContext(ctx context.Context, db *sql.DB, id int) (entities.User, error) {
	row := db.QueryRowContext(ctx, "SELECT id, username, password_hash, last_ram_generated, avatar_ram_id, avatar_box FROM users WHERE id=$1", id)
	user := entities.User{}
	err := row.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.LastRamGenerated, &user.AvatarRamId, &user.AvatarBox)
	if err != nil {
		return entities.User{}, err
	}
	return user, nil
}
