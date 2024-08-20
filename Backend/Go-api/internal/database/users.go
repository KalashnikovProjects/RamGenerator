package database

import (
	"context"
	"database/sql"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/entities"
	_ "github.com/lib/pq"
)

func GetUserContext(ctx context.Context, db SQLQueryExec, id int) (entities.User, error) {
	query := `SELECT id, username, password_hash, daily_ram_generation_time, rams_generated_last_day, clickers_blocked_until, avatar_ram_id, avatar_box FROM users
														WHERE id=$1`
	row := db.QueryRowContext(ctx, query, id)
	user := entities.User{}
	var rawAvatarBox string
	err := row.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.DailyRamGenerationTime, &user.RamsGeneratedLastDay, &user.ClickersBlockedUntil, &user.AvatarRamId, &rawAvatarBox)
	if err != nil {
		return entities.User{}, err
	}
	user.AvatarBox, err = entities.NewBox(rawAvatarBox)
	if err != nil {
		return entities.User{}, err
	}
	return user, nil
}

func GetUserByUsernameContext(ctx context.Context, db SQLQueryExec, username string) (entities.User, error) {
	query := `SELECT id, username, password_hash, daily_ram_generation_time, rams_generated_last_day, clickers_blocked_until, avatar_ram_id, avatar_box FROM users
                                                        WHERE username=$1`
	row := db.QueryRowContext(ctx, query, username)
	user := entities.User{}
	var rawAvatarBox string

	err := row.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.DailyRamGenerationTime, &user.RamsGeneratedLastDay, &user.ClickersBlockedUntil, &user.AvatarRamId, &rawAvatarBox)
	if err != nil {
		return entities.User{}, err
	}
	user.AvatarBox, err = entities.NewBox(rawAvatarBox)
	if err != nil {
		return entities.User{}, err
	}
	return user, nil
}

// CreateUserContext создаёт пользователя и возвращает id
func CreateUserContext(ctx context.Context, db SQLQueryExec, user entities.User) (int, error) {
	avatarBox := ""
	if user.AvatarBox != nil {
		avatarBox = user.AvatarBox.String()
	}
	var id int
	query := `INSERT INTO users (username, password_hash, daily_ram_generation_time, rams_generated_last_day, clickers_blocked_until, avatar_ram_id, avatar_box) 
								VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	err := db.QueryRowContext(ctx, query,
		user.Username, user.PasswordHash, user.DailyRamGenerationTime, user.RamsGeneratedLastDay, &user.ClickersBlockedUntil, user.AvatarRamId, avatarBox).Scan(&id)
	return id, err
}

// UpdateUserContext изменяет поля пользователя, nil или default value поля в переданном объекте игнорируются
func UpdateUserContext(ctx context.Context, db SQLQueryExec, id int, user entities.User) error {
	avatarBox := ""
	if user.AvatarBox != nil {
		avatarBox = user.AvatarBox.String()
	}
	fields := map[string]any{
		"username":                  user.Username,
		"password_hash":             user.PasswordHash,
		"daily_ram_generation_time": user.DailyRamGenerationTime,
		"rams_generated_last_day":   user.RamsGeneratedLastDay,
		"clickers_blocked_until":    user.ClickersBlockedUntil,
		"avatar_ram_id":             user.AvatarRamId,
		"avatar_box":                avatarBox,
	}
	query, args := GenerateQueryAndArgsForUpdate("users", fields, "id=$1", id)
	_, err := db.ExecContext(ctx, query, args...)
	return err
}

// UpdateUserByUsernameContext изменяет поля пользователя, nil или default value поля в переданном объекте игнорируются
func UpdateUserByUsernameContext(ctx context.Context, db SQLQueryExec, username string, user entities.User) error {
	avatarBox := ""
	if user.AvatarBox != nil {
		avatarBox = user.AvatarBox.String()
	}
	fields := map[string]any{
		"username":                  user.Username,
		"password_hash":             user.PasswordHash,
		"daily_ram_generation_time": user.DailyRamGenerationTime,
		"rams_generated_last_day":   user.RamsGeneratedLastDay,
		"clickers_blocked_until":    user.ClickersBlockedUntil,
		"avatar_ram_id":             user.AvatarRamId,
		"avatar_box":                avatarBox,
	}
	query, args := GenerateQueryAndArgsForUpdate("users", fields, "username=$1", username)
	_, err := db.ExecContext(ctx, query, args...)
	return err
}

func DeleteUserRamsContext(ctx context.Context, db SQLQueryExec, id int) error {
	query := `DELETE FROM rams WHERE user_id=$1`
	_, err := db.ExecContext(ctx, query, id)
	return err
}

func DeleteUserContext(ctx context.Context, db SQLQueryExec, id int) error {
	err := DeleteUserRamsContext(ctx, db, id)
	if err != nil {
		return err
	}
	query := `DELETE FROM users WHERE id=$1`
	_, err = db.ExecContext(ctx, query, id)
	return err
}

func DeleteUserRamsByUsernameContext(ctx context.Context, db SQLQueryExec, username string) error {
	query := `DELETE FROM rams AS R
			USING users AS U
            WHERE R.user_id=u.id AND u.username=$1`
	_, err := db.ExecContext(ctx, query, username)
	return err
}

func DeleteUserByUsernameContext(ctx context.Context, db SQLTXQueryExec, username string) error {
	tx, _ := db.BeginTx(ctx, &sql.TxOptions{})
	err := DeleteUserRamsByUsernameContext(ctx, tx, username)
	if err != nil {
		tx.Rollback()
		return err
	}
	query := `DELETE FROM users WHERE username=$1`
	_, err = tx.ExecContext(ctx, query, username)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func GetUserByRamIdContext(ctx context.Context, db SQLQueryExec, id int) (entities.User, error) {
	ram, err := GetRamContext(ctx, db, id)
	if err != nil {
		return entities.User{}, err
	}
	return GetUserContext(ctx, db, ram.UserId)
}

func UpdateUserClickersBlockedIfCan(ctx context.Context, db SQLQueryExec, id, cantGenerateRamUntil, nowTime int) error {
	query := "UPDATE users SET clickers_blocked_until=$1 WHERE id=$2 AND clickers_blocked_until<$3 RETURNING id"
	row := db.QueryRowContext(ctx, query, cantGenerateRamUntil, id, nowTime)
	var dbId int
	return row.Scan(&dbId)
}

func UpdateUserClickersBlockedToZero(ctx context.Context, db SQLQueryExec, id int) error {
	query := "UPDATE users SET clickers_blocked_until=0 WHERE id=$1"
	_, err := db.ExecContext(ctx, query, id)
	return err
}
