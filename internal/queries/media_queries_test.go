package queries_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/test"
)

func TestCreateMedia(t *testing.T) {
	test.Short(t)
ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "test@example.com",
		})
		if err != nil {
			return err
		}
		type args struct {
			ctx   context.Context
			exec  database.Dbx
			media *models.Medium
		}
		tests := []struct {
			name    string
			args    args
			want    *models.Medium
			wantErr bool
		}{
			{
				name: "Create media",
				args: args{
					ctx:  ctx,
					exec: dbxx,
					media: &models.Medium{
						UserID:           &user.ID,
						Disk:             "local",
						Directory:        "test/dir",
						Filename:         "test.jpg",
						OriginalFilename: "test_original.jpg",
						Extension:        "jpg",
						MimeType:         "image/jpeg",
						Size:             1024,
					},
				},
				want: &models.Medium{
					UserID:           &user.ID,
					Disk:             "local",
					Directory:        "test/dir",
					Filename:         "test.jpg",
					OriginalFilename: "test_original.jpg",
					Extension:        "jpg",
					MimeType:         "image/jpeg",
					Size:             1024,
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.CreateMedia(tt.args.ctx, tt.args.exec, tt.args.media)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateMedia() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got.Disk, tt.want.Disk) {
					t.Errorf("CreateMedia() = %v, want %v", got.Disk, tt.want.Disk)
				}
				if !reflect.DeepEqual(got.Directory, tt.want.Directory) {
					t.Errorf("CreateMedia() = %v, want %v", got.Directory, tt.want.Directory)
				}
				if !reflect.DeepEqual(got.Filename, tt.want.Filename) {
					t.Errorf("CreateMedia() = %v, want %v", got.Filename, tt.want.Filename)
				}
				if !reflect.DeepEqual(got.OriginalFilename, tt.want.OriginalFilename) {
					t.Errorf("CreateMedia() = %v, want %v", got.OriginalFilename, tt.want.OriginalFilename)
				}
				if !reflect.DeepEqual(got.Extension, tt.want.Extension) {
					t.Errorf("CreateMedia() = %v, want %v", got.Extension, tt.want.Extension)
				}
				if !reflect.DeepEqual(got.MimeType, tt.want.MimeType) {
					t.Errorf("CreateMedia() = %v, want %v", got.MimeType, tt.want.MimeType)
				}
				if !reflect.DeepEqual(got.Size, tt.want.Size) {
					t.Errorf("CreateMedia() = %v, want %v", got.Size, tt.want.Size)
				}
				if !reflect.DeepEqual(got.UserID, tt.want.UserID) {
					t.Errorf("CreateMedia() = %v, want %v", got.UserID, tt.want.UserID)
				}
			})
		}
		return test.ErrEndTest
	})
}

func TestFindMediaByID(t *testing.T) {
	test.Short(t)
ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "test@example.com",
		})
		if err != nil {
			return err
		}

		media, err := queries.CreateMedia(ctx, dbxx, &models.Medium{
			UserID:           &user.ID,
			Disk:             "local",
			Directory:        "test/dir",
			Filename:         "test.jpg",
			OriginalFilename: "test_original.jpg",
			Extension:        "jpg",
			MimeType:         "image/jpeg",
			Size:             1024,
		})
		if err != nil {
			return err
		}

		type args struct {
			ctx  context.Context
			exec database.Dbx
			id   uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			want    *models.Medium
			wantErr bool
		}{
			{
				name: "Find existing media",
				args: args{
					ctx:  ctx,
					exec: dbxx,
					id:   media.ID,
				},
				want:    media,
				wantErr: false,
			},
			{
				name: "Find non-existing media",
				args: args{
					ctx:  ctx,
					exec: dbxx,
					id:   uuid.New(),
				},
				want:    nil,
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.FindMediaByID(tt.args.ctx, tt.args.exec, tt.args.id)
				if (err != nil) != tt.wantErr {
					t.Errorf("FindMediaByID() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.want == nil {
					if got != nil {
						t.Errorf("FindMediaByID() = %v, want nil", got)
					}
					return
				}
				if !reflect.DeepEqual(got.ID, tt.want.ID) {
					t.Errorf("FindMediaByID() got ID = %v, want %v", got.ID, tt.want.ID)
				}
				if !reflect.DeepEqual(got.UserID, tt.want.UserID) {
					t.Errorf("FindMediaByID() got UserID = %v, want %v", got.UserID, tt.want.UserID)
				}
				if !reflect.DeepEqual(got.Disk, tt.want.Disk) {
					t.Errorf("FindMediaByID() got Disk = %v, want %v", got.Disk, tt.want.Disk)
				}
			})
		}
		return test.ErrEndTest
	})
}
