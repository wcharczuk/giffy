package initialize

import (
	"fmt"

	"github.com/blend/go-sdk/db/migration"
	"github.com/wcharczuk/giffy/server/config"
)

// Initialize returns the initialize migrations.
func Initialize(cfg *config.Giffy) migration.Migration {

	return migration.New(
		migration.New(
			migration.NewStep(
				migration.AlwaysRun(),
				migration.Statements(
					"DROP SCHEMA public CASCADE;",
					"CREATE SCHEMA public;",
				),
			).WithLabel("recreate schema"),
		),

		migration.New(
			migration.NewStep(
				migration.AlwaysRun(),
				migration.Statements(
					"create extension if not exists pgcrypto;",
					"create extension if not exists pg_trgm;",
				),
			).WithLabel("add pg_trgm and pgcrypto"),
		),

		migration.New(
			migration.NewStep(
				migration.TableNotExists("error"),
				migration.Statements(
					`CREATE TABLE error (
						uuid varchar(32) not null,
						created_utc timestamp not null,
						message varchar(255) not null,
						stack_trace varchar(1024),

						verb varchar(8),
						proto varchar(8),
						host varchar(255),
						path varchar(255),
						query varchar(255)
					);`,
					`ALTER TABLE error ADD CONSTRAINT pk_error_uuid PRIMARY KEY (uuid);`,
					`CREATE INDEX ix_error_created_utc ON error(created_utc);`,
				),
			),
		),

		migration.New(
			migration.NewStep(
				migration.TableNotExists("users"),
				migration.Statements(
					`CREATE TABLE users (
						id serial not null,
						uuid varchar(32) not null,
						created_utc timestamp not null,
						username varchar(255) not null,
						first_name varchar(64),
						last_name varchar(64),
						email_address varchar(255),
						is_email_verified boolean not null,
						is_admin boolean not null default false,
						is_moderator boolean not null default false,
						is_banned boolean not null default false
					);`,
					`ALTER TABLE users ADD CONSTRAINT pk_users_id PRIMARY KEY (id);`,
					`ALTER TABLE users ADD CONSTRAINT uk_users_uuid UNIQUE (uuid);`,
					`ALTER TABLE users ADD CONSTRAINT uk_users_username UNIQUE (username);`,
				),
			),
		),

		migration.New(
			migration.NewStep(
				migration.TableNotExists("user_auth"),
				migration.Statements(
					`CREATE TABLE user_auth (
						user_id bigint not null,
						provider varchar(32) not null,
						timestamp_utc timestamp not null,
						auth_token bytea not null,
						auth_token_hash bytea not null,
						auth_secret bytea
					);`,
					`ALTER TABLE user_auth ADD CONSTRAINT pk_user_auth_user_id_provider PRIMARY KEY (user_id,provider);`,
					`ALTER TABLE user_auth ADD CONSTRAINT fk_user_auth_user_id FOREIGN KEY (user_id) REFERENCES users(id);`,
					`CREATE INDEX ix_user_auth_auth_token_hash ON user_auth(auth_token_hash);`,
				),
			),
		),

		migration.New(
			migration.NewStep(
				migration.TableNotExists("user_session"),
				migration.Statements(
					`CREATE TABLE user_session (
						session_id varchar(32) not null,
						user_id bigint not null,
						timestamp_utc timestamp not null
					);`,
					`ALTER TABLE user_session ADD CONSTRAINT pk_user_session_session_id PRIMARY KEY (session_id);`,
					`ALTER TABLE user_session ADD CONSTRAINT fk_user_session_user_id FOREIGN KEY (user_id) REFERENCES users(id);`,
				),
			),
		),

		migration.New(
			migration.NewStep(
				migration.TableNotExists("content_rating"),
				migration.Statements(
					`CREATE TABLE content_rating (
						id int not null,
						name varchar(32) not null,
						description varchar(1024) not null
					);`,
					`ALTER TABLE content_rating ADD CONSTRAINT pk_content_rating_id PRIMARY KEY (id);`,
				),
			),
		),

		migration.New(
			migration.NewStep(
				migration.TableNotExists("image"),
				migration.Statements(
					`CREATE TABLE image (
						id serial not null,
						uuid varchar(32) not null,
						created_utc timestamp not null,
						created_by bigint not null,

						display_name varchar(255),

						md5 bytea not null,
						s3_read_url varchar(1024),
						s3_bucket varchar(64) not null,
						s3_key varchar(64) not null,

						content_rating int not null default 3,

						width int not null,
						height int not null,
						file_size int not null,

						extension varchar(8)
					);`,
					`ALTER TABLE image ADD CONSTRAINT pk_image_id PRIMARY KEY (id);`,
					`ALTER TABLE image ADD CONSTRAINT uk_image_uuid UNIQUE (uuid);`,
					`ALTER TABLE image ADD CONSTRAINT uk_image_md5 UNIQUE (md5);`,
					`ALTER TABLE image ADD CONSTRAINT fk_image_created_by_user_id FOREIGN KEY (created_by) REFERENCES users(id);`,
					`ALTER TABLE image ADD CONSTRAINT fk_image_content_rating FOREIGN KEY (content_rating) REFERENCES content_rating(id);`,
				),
			),
		),

		migration.New(
			migration.NewStep(
				migration.TableNotExists("tag"),
				migration.Statements(
					`CREATE TABLE tag (
						id serial not null,
						uuid varchar(32) not null,
						created_utc timestamp not null,
						created_by bigint not null,
						tag_value varchar(32) not null
					);`,
					`ALTER TABLE tag ADD CONSTRAINT pk_tag_id PRIMARY KEY (id);`,
					`ALTER TABLE tag ADD CONSTRAINT uk_tag_uuid UNIQUE (uuid);`,
					`ALTER TABLE tag ADD CONSTRAINT uk_tag_tag_value UNIQUE (tag_value);`,
					`ALTER TABLE tag ADD CONSTRAINT fk_tag_created_by_user_id FOREIGN KEY (created_by) REFERENCES users(id);`,
					`CREATE INDEX idx_tag_tag_value_gin ON tag USING gin (tag_value gin_trgm_ops);`,
				),
			),
		),

		migration.New(
			migration.NewStep(
				migration.TableNotExists("vote_summary"),
				migration.Statements(
					`CREATE TABLE vote_summary (
						image_id bigint not null,
						tag_id bigint not null,
						created_utc timestamp not null,
						last_vote_by bigint not null,
						last_vote_utc timestamp not null,
						votes_for int not null,
						votes_against int not null,
						votes_total int not null
					);`,
					`ALTER TABLE vote_summary ADD CONSTRAINT pk_vote_summary_image_id_tag_id PRIMARY KEY (image_id, tag_id);`,
					`ALTER TABLE vote_summary ADD CONSTRAINT fk_vote_summary_image_id FOREIGN KEY (image_id) REFERENCES image(id);`,
					`ALTER TABLE vote_summary ADD CONSTRAINT fk_vote_summary_tag_id FOREIGN KEY (tag_id) REFERENCES tag(id);`,
				),
			),
		),

		migration.New(
			migration.NewStep(
				migration.TableNotExists("vote"),
				migration.Statements(
					`CREATE TABLE vote (
						user_id bigint not null,
						image_id bigint not null,
						tag_id bigint not null,
						created_utc timestamp not null,
						is_upvote bool not null
					);`,
					`ALTER TABLE vote ADD CONSTRAINT pk_vote_user_id_image_id_tag_id PRIMARY KEY (user_id, image_id, tag_id);`,
					`ALTER TABLE vote ADD CONSTRAINT fk_vote_user_id FOREIGN KEY (user_id) REFERENCES users(id);`,
					`ALTER TABLE vote ADD CONSTRAINT fk_vote_image_id FOREIGN KEY (image_id) REFERENCES image(id);`,
					`ALTER TABLE vote ADD CONSTRAINT fk_vote_tag_id FOREIGN KEY (tag_id) REFERENCES tag(id);`,
				),
			),
		),

		migration.New(
			migration.NewStep(
				migration.TableNotExists("moderation"),
				migration.Statements(
					`CREATE TABLE moderation (
						user_id bigint not null,
						uuid varchar(32) not null,
						timestamp_utc timestamp not null,
						verb varchar(32) not null,
						object varchar(255) not null,
						noun varchar(255),
						secondary_noun varchar(255)
					);`,
					`ALTER TABLE moderation ADD CONSTRAINT pk_moderation_uuid PRIMARY KEY (uuid);`,
					`ALTER TABLE moderation add CONSTRAINT fk_moderation_user_id FOREIGN KEY (user_id) REFERENCES users(id);`,
				),
			),
		),

		migration.New(
			migration.NewStep(
				migration.TableNotExists("search_history"),
				migration.Statements(
					`create table search_history (
						source varchar(255) not null,

						source_team_identifier varchar(255),
						source_channel_identifier varchar(255),
						source_user_identifier varchar(255),

						source_team_name varchar(255),
						source_channel_name varchar(255),
						source_user_name varchar(255),

						timestamp_utc timestamp not null,
						search_query varchar(255) not null,

						did_find_match boolean not null,

						image_id bigint,
						tag_id bigint
					);`,
				),
			),
		),

		migration.New(
			migration.NewStep(
				migration.TableNotExists("slack_team"),
				migration.Statements(
					`CREATE TABLE slack_team (
						team_id varchar(32) not null,
						team_name varchar(128) not null,
						created_utc timestamp not null,
						is_enabled bool not null,
						created_by_id varchar(32) not null,
						created_by_name varchar(128) not null,
						content_rating int not null
					);`,
					`ALTER TABLE slack_team ADD CONSTRAINT pk_content_rating_team_id PRIMARY KEY (team_id);`,
				),
			),
		),

		migration.New(
			migration.NewStep(
				migration.AlwaysRun(),
				migration.Statements(
					fmt.Sprintf(`
						insert into users
							(uuid, username, created_utc, first_name, last_name, email_address, is_email_verified, is_admin, is_moderator)
						select
							'a68aac8196e444d4a3e570192a20f369', '%s', current_timestamp, 'Will', 'Charczuk', '%s', true, true, true
						where not exists (select 1 from users where uuid = 'a68aac8196e444d4a3e570192a20f369');
						`,
						cfg.GetAdminUserEmail(), cfg.GetAdminUserEmail()),
				),
			).WithLabel("create admin user"),
		),

		migration.New(
			migration.NewStep(
				migration.AlwaysRun(),
				migration.Statements(
					"INSERT INTO content_rating (id, name, description) select 1, 'G', 'General Audiences; no violence or sexual content. No live action.' where not exists ( select 1 from content_rating where id = 1 );",
					"INSERT INTO content_rating (id, name, description) select 2, 'PG', 'Parental Guidance; limited violence and sexual content. Some live action' where not exists ( select 1 from content_rating where id = 2 );",
					"INSERT INTO content_rating (id, name, description) select 3, 'PG-13', 'Parental Guidance (13 and over); some violence and sexual content. Live action and animated.' where not exists ( select 1 from content_rating where id = 3 );",
					"INSERT INTO content_rating (id, name, description) select 4, 'R', 'Restricted; very violent or sexual in content.' where not exists ( select 1 from content_rating where id = 4 );",
					"INSERT INTO content_rating (id, name, description) select 5, 'NR', 'Not Rated; reserved for the dankest of may-mays, may be disturbing. Usually NSFW, will generally get you fired if you look at these at work.' where not exists ( select 1 from content_rating where id = 5 );",
				),
			).WithLabel("create content ratings"),
		),
	)
}
