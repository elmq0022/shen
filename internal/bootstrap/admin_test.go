package bootstrap_test

import (
	"os"
	"testing"

	"github.com/elmq0022/shen/internal/bootstrap"
)

func TestGetAdminUser(t *testing.T) {
	t.Run("returns default when env var not set", func(t *testing.T) {
		os.Unsetenv(bootstrap.AdminUserEnv)

		result := bootstrap.GetAdminUser()

		if result != bootstrap.DefaultAdminUser {
			t.Errorf("expected %q, got %q", bootstrap.DefaultAdminUser, result)
		}
	})

	t.Run("returns env var value when set", func(t *testing.T) {
		expected := "custom_admin"
		os.Setenv(bootstrap.AdminUserEnv, expected)
		t.Cleanup(func() { os.Unsetenv(bootstrap.AdminUserEnv) })

		result := bootstrap.GetAdminUser()

		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("returns empty string when env var is empty", func(t *testing.T) {
		os.Setenv(bootstrap.AdminUserEnv, "")
		t.Cleanup(func() { os.Unsetenv(bootstrap.AdminUserEnv) })

		result := bootstrap.GetAdminUser()

		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
	})
}

func TestGetAdminPassword(t *testing.T) {
	t.Run("returns default when env var not set", func(t *testing.T) {
		os.Unsetenv(bootstrap.AdminPasswordEnv)

		result := bootstrap.GetAdminPassword()

		if result != bootstrap.DefaultAdminPassword {
			t.Errorf("expected %q, got %q", bootstrap.DefaultAdminPassword, result)
		}
	})

	t.Run("returns env var value when set", func(t *testing.T) {
		expected := "super_secret_password"
		os.Setenv(bootstrap.AdminPasswordEnv, expected)
		t.Cleanup(func() { os.Unsetenv(bootstrap.AdminPasswordEnv) })

		result := bootstrap.GetAdminPassword()

		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("returns empty string when env var is empty", func(t *testing.T) {
		os.Setenv(bootstrap.AdminPasswordEnv, "")
		t.Cleanup(func() { os.Unsetenv(bootstrap.AdminPasswordEnv) })

		result := bootstrap.GetAdminPassword()

		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
	})
}
