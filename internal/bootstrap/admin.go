package bootstrap

import (
	"context"
	"os"

	"github.com/elmq0022/shen/internal/auth"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	DefaultAdminUser     string = "admin"
	DefaultAdminPassword string = "admin"

	AdminUserEnv     string = "SHEN_ADMIN_USERNAME"
	AdminPasswordEnv string = "SHEN_ADMIN_PASSWORD"
)

func GetAdminUser() string {
	user, exists := os.LookupEnv(AdminUserEnv)
	if exists {
		return user
	}
	return DefaultAdminUser
}

func GetAdminPassword() string {
	password, exists := os.LookupEnv(AdminPasswordEnv)
	if exists {
		return password
	}
	return DefaultAdminPassword
}

func CreateAdmin(ctx context.Context, queries *db.Queries, hashPassword auth.HashFunc) {
	numUser, err := queries.CountUsers(ctx)
	if err != nil {
		panic(err)
	}
	if numUser > 0 {
		return
	}

	user := GetAdminUser()
	password := GetAdminPassword()
	role, err := queries.GetRoleByName(ctx, "admin")
	if err != nil {
		panic(err)
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		panic(err)
	}

	_, err = queries.CreateUser(ctx, db.CreateUserParams{
		Username: user,
		HashedPassword: pgtype.Text{
			String: hashedPassword,
			Valid:  true,
		},
		Role: role.ID,
	})
	if err != nil {
		panic(err)
	}
}
