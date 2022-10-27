package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-cmp/cmp"
	"github.com/jinchi2013/service/busniess/data/store/user"
	"github.com/jinchi2013/service/busniess/data/tests"
	"github.com/jinchi2013/service/busniess/sys/auth"
	"github.com/jinchi2013/service/busniess/sys/database"
)

var dbc = tests.DBContainer{
	Image: "postgres:13-alpine",
	Port:  "5432",
	Args:  []string{"-e", "POSTGRES_PASSWORK=postgres"},
}

func TestUser(t *testing.T) {
	log, db, teardown := tests.NewUnit(t, dbc)
	t.Cleanup(teardown)

	// new user store
	store := user.NewStore(log, db)

	t.Log("Given the need to work with User records")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single User.", testID)
		{
			ctx := context.Background()
			now := time.Date(2022, time.October, 1, 0, 0, 0, 0, time.UTC)

			nu := user.NewUser{
				Name:            "CHI JIN",
				Email:           "chi@service3go.com",
				Roles:           []string{auth.RoleAdmin},
				Password:        "gophers",
				PasswordConfirm: "gophers",
			}

			usr, err := store.Create(ctx, nu, now)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create user : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%sTest %d:\tShould be able to create user.", tests.Success, testID)

			claims := auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Issuer:    "service project",
					Subject:   usr.ID,
					ExpiresAt: time.Now().Add(time.Hour).Unix(),
					IssuedAt:  time.Now().UTC().Unix(),
				},
				Roles: []string{auth.RoleUser},
			}

			saved, err := store.QueryByID(ctx, claims, usr.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve user by ID: %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%sTest %d:\tShould be able to retrieve user by ID.", tests.Success, testID)

			if diff := cmp.Diff(usr, saved); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same user. Diff:\n%s", tests.Failed, testID, err)
			}

			t.Logf("\t%sTest %d:\tShould get back the same user.", tests.Success, testID)

			upd := user.UpdateUser{
				Name:  tests.StringPointer("QIfang, CHI"),
				Email: tests.StringPointer("qifang@119village.com"),
			}

			claims = auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Issuer:    "service project",
					ExpiresAt: time.Now().Add(time.Hour).Unix(),
					IssuedAt:  time.Now().UTC().Unix(),
				},
				Roles: []string{auth.RoleAdmin},
			}

			if err := store.Update(ctx, claims, usr.ID, upd, now); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to update user %s.", tests.Failed, testID, err)
			}

			t.Logf("\t%s\tTest %d:\tShould be able to update user.", tests.Failed, testID)

			saved, err = store.QueryByEmail(ctx, claims, *upd.Email)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve user by Email : %s.", tests.Failed, testID, err)
			}

			t.Logf("\t%s\tTest %d:\tShould be able to retrieve user by Email.", tests.Failed, testID)

			if saved.Name != *upd.Name {
				t.Errorf("\t%s\tTest %d:\tShould be able to see the updates to Name.", tests.Failed, testID)
				t.Logf("\t\tTest %d:\tGot: %v", testID, saved.Name)
				t.Logf("\t\tTest %d:\tExp: %v", testID, *upd.Name)
			} else {
				t.Logf("\t%s\tTest %d:\tShould be able to see updates to Name.", tests.Success, testID)
			}

			if saved.Email != *upd.Email {
				t.Errorf("\t%s\tTest %d:\tShould be able to see the updates to Email.", tests.Failed, testID)
				t.Logf("\t\tTest %d:\tGot: %v", testID, saved.Email)
				t.Logf("\t\tTest %d:\tExp: %v", testID, *upd.Email)
			} else {
				t.Logf("\t%s\tTest %d:\tShould be able to see updates to Email.", tests.Success, testID)
			}

			if err := store.Delete(ctx, claims, usr.ID); err != nil {
				t.Fatalf("\t%sTest %d:\tShould be able to delete user : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to delete user.", tests.Success, testID)

			_, err = store.QueryByID(ctx, claims, usr.ID)
			if !errors.Is(err, database.ErrNotFound) {
				t.Fatalf("\t%sTest %d:\tShould NOT be able to retrieve user : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould NOT be able to retrieve user", tests.Success, testID)
		}
	}
}
