package models

import (
	"time"

	"github.com/google/uuid"
)

type BlogPostStatus string

const (
	BlogPostStatusDraft     BlogPostStatus = "draft"
	BlogPostStatusPublished BlogPostStatus = "published"
	BlogPostStatusArchived  BlogPostStatus = "archived"
)

type BlogContentFormat string

const (
	BlogContentFormatTiptap   BlogContentFormat = "tiptap"
	BlogContentFormatMarkdown BlogContentFormat = "markdown"
)

type BlogPost struct {
	_                  struct{}          `db:"posts" schema:"blog" json:"-"`
	ID                 uuid.UUID         `db:"id" json:"id"`
	Slug               string            `db:"slug" json:"slug"`
	Title              string            `db:"title" json:"title"`
	Content            string            `db:"content" json:"content"`
	ContentFormat      BlogContentFormat `db:"content_format" json:"content_format"`
	Status             BlogPostStatus    `db:"status" json:"status"`
	AuthorID           uuid.UUID         `db:"author_id" json:"author_id"`
	PublishedAt        *time.Time        `db:"published_at" json:"published_at" nullable:"true"`
	FeaturedImageKey   *string           `db:"featured_image_key" json:"featured_image_key" nullable:"true"`
	SeoTitle           *string           `db:"seo_title" json:"seo_title" nullable:"true"`
	SeoDescription     *string           `db:"seo_description" json:"seo_description" nullable:"true"`
	ReadingTimeMinutes *int              `db:"reading_time_minutes" json:"reading_time_minutes" nullable:"true"`
	ViewCount          int64             `db:"view_count" json:"view_count"`
	CreatedAt          time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time         `db:"updated_at" json:"updated_at"`
	Author             *User             `db:"author" src:"author_id" dest:"id" table:"auth.users" json:"author,omitempty"`
	Tags               []*BlogTag        `db:"tags" src:"id" dest:"id" table:"blog.tags" through:"blog.post_tags" through_src:"post_id" through_dest:"tag_id" json:"tags,omitempty"`
}

type BlogTag struct {
	_         struct{}  `db:"tags" schema:"blog" json:"-"`
	ID        uuid.UUID `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Slug      string    `db:"slug" json:"slug"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
