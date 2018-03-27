package migrations

import "github.com/blendlabs/spiffy/migration"

// Migrations returns the migrations.
func Migrations() migration.Migration {
	return migration.New(
		contentRating(),
		slackTeam(),
		errors(),
		sessionIDs(),
	)
}

func contentRating() migration.Migration {
	createContentRating := migration.NewStep(
		migration.TableNotExists("content_rating"),
		migration.Statements(`
				CREATE TABLE content_rating (
					id int not null,
					name varchar(32) not null,
					description varchar(1024) not null
				);`,
			`ALTER TABLE content_rating ADD CONSTRAINT pk_content_rating_id PRIMARY KEY (id);`,
			`INSERT INTO content_rating (id, name, description) VALUES (1, 'G', 'General Audiences; no violence or sexual content. No live action.');`,
			`INSERT INTO content_rating (id, name, description) VALUES (2, 'PG', 'Parental Guidance; limited violence and sexual content. Some live action');`,
			`INSERT INTO content_rating (id, name, description) VALUES (3, 'PG-13', 'Parental Guidance (13 and over); some violence and sexual content. Live action and animated.');`,
			`INSERT INTO content_rating (id, name, description) VALUES (4, 'R', 'Restricted; very violent or sexual in content.');`,
			`INSERT INTO content_rating (id, name, description) VALUES (5, 'NR', 'Not Rated; reserved for the dankest of may-mays, may be disturbing. Usually NSFW, will generally get you fired if you look at these at work.');`,
		),
	)

	addContentRatingToImage := migration.NewStep(
		migration.ColumnNotExists("image", "content_rating"),
		migration.Statements(
			"ALTER TABLE image ADD content_rating int;",
			"UPDATE image set content_rating = 3;",
			"UPDATE image set content_rating = 5 WHERE is_censored = true;",
		),
	)

	dropIsCensoredFromImage := migration.NewStep(
		migration.ColumnExists("image", "is_censored"),
		migration.Statements(
			"ALTER TABLE image DROP is_censored;",
		),
	)

	return migration.New(
		createContentRating,
		addContentRatingToImage,
		dropIsCensoredFromImage,
	)
}

func slackTeam() migration.Migration {
	return migration.New(
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
				`INSERT INTO slack_team (team_id, team_name, created_utc, is_enabled, created_by_id, created_by_name, content_rating)
				SELECT
					team_id
					, team_name
					, created_utc
					, true as is_enabled
					, created_by_id
					, created_by_name
					, 3 as content_rating
				FROM (
					SELECT
						source_team_identifier as team_id
						, source_team_name as team_name
						, current_date as created_utc
						, source_user_identifier as created_by_id
						, source_user_name as created_by_name
						, ROW_NUMBER() OVER(partition by source_team_identifier order by timestamp_utc asc) as team_rank
					FROM
						search_history
					WHERE
						source_team_identifier is not null
						and length(source_team_identifier) > 0
				) as data
				where data.team_rank = 1`,
			),
		),
	)
}

func errors() migration.Migration {
	return migration.New(
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
	)
}

func sessionIDs() migration.Migration {
	return migration.New(
		migration.NewStep(
			migration.ColumnExists("user_session", "session_id"),
			migration.Statements(`AlTER TABLE user_session ALTER session_id TYPE varchar(128)`),
		),
	)
}
