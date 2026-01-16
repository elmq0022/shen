package integration

import "testing"

func TestMileStone2_UserAdministration(t *testing.T) {
	resetDB(t)
	startTestServer(t, shen)
	// := setTempXDG(t)

	t.Run("admin creates new user", func(t *testing.T) {
		t.Skip()
	})

	t.Run("admin lists users", func(t *testing.T) {
		t.Skip()
	})

	t.Run("admin updates user role", func(t *testing.T) {
		t.Skip()
	})

	t.Run("user updates own password", func(t *testing.T) {
		t.Skip()
	})

	t.Run("admin updates another users' password", func(t *testing.T) {
		t.Skip()
	})

	t.Run("user cannot update another user's password", func(t *testing.T) {
		t.Skip()
	})

	t.Run("admin soft deletes user", func(t *testing.T) {
		t.Skip()
	})

	t.Run("non-admin cannot access user management endpoints", func(t *testing.T) {
		t.Skip()
	})

	t.Run("service accounts cannot login to shen", func(t *testing.T) {
		t.Skip()
	})
}
