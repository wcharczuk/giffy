package migration

import (
	"github.com/blendlabs/spiffy"
	"github.com/blendlabs/spiffy/migration"
)

/*
begin;

select migrate_create_table('content_rating', '
	CREATE TABLE content_rating (
		id int not null,
		name varchar(32) not null,
		description varchar(1024) not null
	);
	ALTER TABLE content_rating ADD CONSTRAINT pk_content_rating_id PRIMARY KEY (id);

	INSERT INTO content_rating (id, name) VALUES (1, ''G'', ''General Audiences; no violence or sexual content. No live action.'');
	INSERT INTO content_rating (id, name) VALUES (2, ''PG'', ''Parental Guidance; limited violence and sexual content. Some live action'');
	INSERT INTO content_rating (id, name) VALUES (3, ''PG-13'', ''Parental Guidance (13 and over); some violence and sexual content. Live action and animated.'');
	INSERT INTO content_rating (id, name) VALUES (4, ''R'', ''Restricted; very violent or sexual in content.'');
	INSERT INTO content_rating (id, name) VALUES (5, ''NR'', ''Not Rated; reserved for the dankest of may-mays, may be disturbing. Usually NSFW, will generally get you fired if you look at these at work.'');
');

select migrate_create_column('image', 'content_rating', '
	ALTER TABLE image ADD content_rating int;

	UPDATE image set content_rating = 3;
	UPDATE image set content_rating = 5 WHERE is_censored = true;
');

select migrate_alter_column('image', 'is_censored', '
	ALTER TABLE image DROP is_censored;
');

commit;
*/

func addContentRating() migration.Migration {
	return migration.New(
		"adds `content_rating` table",
		migration.Step(
			migration.CreateTable,
			migration.Body(`
				CREATE TABLE content_rating (
					id int not null,
					name varchar(32) not null,
					description varchar(1024) not null
				);`,
				`ALTER TABLE content_rating ADD CONSTRAINT pk_content_rating_id PRIMARY KEY (id);`,
				`INSERT INTO content_rating (id, name) VALUES (1, ''G'', ''General Audiences; no violence or sexual content. No live action.'');`,
				`INSERT INTO content_rating (id, name) VALUES (2, ''PG'', ''Parental Guidance; limited violence and sexual content. Some live action'');`,
				`INSERT INTO content_rating (id, name) VALUES (3, ''PG-13'', ''Parental Guidance (13 and over); some violence and sexual content. Live action and animated.'');`,
				`INSERT INTO content_rating (id, name) VALUES (4, ''R'', ''Restricted; very violent or sexual in content.'');`,
				`INSERT INTO content_rating (id, name) VALUES (5, ''NR'', ''Not Rated; reserved for the dankest of may-mays, may be disturbing. Usually NSFW, will generally get you fired if you look at these at work.'');`,
			),
			"content_rating",
		),
	)
}

// Run applies all migrations
func Run() error {
	return migration.Default(func(suite migration.Migration) error {
		return suite.Apply(spiffy.DefaultDb())
	})
}
