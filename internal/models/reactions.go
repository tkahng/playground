package models

import (
	"time"

	"github.com/google/uuid"
)

type UserReaction struct {
	// id uuid not null primary key default gen_random_uuid(),
	// user_id uuid references public.users on delete cascade on update cascade,
	// type text not null,
	// reaction text,
	// ip_address text,
	// country text,
	// city text,
	// metadata jsonb,
	// created_at timestamptz not null default now(),
	// updated_at timestamptz not null default now()
	_         struct{}   `db:"user_reactions" json:"-"`
	ID        uuid.UUID  `db:"id" json:"id"`
	UserID    *uuid.UUID `db:"user_id" json:"user_id"`
	Type      string     `db:"type" json:"type"`
	Reaction  *string    `db:"otp" json:"otp"`
	IpAddress *string    `db:"ip_address" json:"ip_address"`
	Country   *string    `db:"country" json:"country"`
	City      *string    `db:"city" json:"city"`
	Metadata  []byte     `db:"metadata" json:"metadata"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	User      *User      `db:"users" src:"user_id" dest:"id" table:"users" json:"user,omitempty"`
}
