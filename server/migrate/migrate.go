package migrate

import (
	"github.com/blendlabs/spiffy"
	"github.com/blendlabs/spiffy/migration"
)

func contentRating() migration.Migration {
	createContentRating := migration.Step(
		migration.CreateTable,
		migration.Body(`
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
		"content_rating",
	)

	addContentRatingToImage := migration.Step(
		migration.CreateColumn,
		migration.Body(
			"ALTER TABLE image ADD content_rating int;",
			"UPDATE image set content_rating = 3;",
			"UPDATE image set content_rating = 5 WHERE is_censored = true;",
		),
		"image",
		"content_rating",
	)

	dropIsCensoredFromImage := migration.Step(
		migration.AlterColumn,
		migration.Body(
			"ALTER TABLE image DROP is_censored;",
		),
		"image",
		"is_censored",
	)

	return migration.New(
		"adds `content_rating`",
		createContentRating,
		addContentRatingToImage,
		dropIsCensoredFromImage,
	)
}

// Register register migrations
func Register() {
	migration.Register(contentRating())
}

// Run applies all migrations
func Run() error {
	return migration.Default(func(suite migration.Migration) error {
		suite.Logged(migration.NewLogger())
		return suite.Apply(spiffy.DefaultDb())
	})
}
