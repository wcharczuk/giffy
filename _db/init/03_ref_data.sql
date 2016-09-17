insert into users (uuid, username, created_utc, first_name, last_name, email_address, is_email_verified, is_admin, is_moderator)
values
('a68aac8196e444d4a3e570192a20f369', 'will.charczuk@gmail.com', current_timestamp, 'Will', 'Charczuk', 'will.charczuk@gmail.com', true, true, true);

INSERT INTO content_rating (id, name, description) VALUES (1, 'G', 'General Audiences; no violence or sexual content. No live action.');
INSERT INTO content_rating (id, name, description) VALUES (2, 'PG', 'Parental Guidance; limited violence and sexual content. Some live action');
INSERT INTO content_rating (id, name, description) VALUES (3, 'PG-13', 'Parental Guidance (13 and over); some violence and sexual content. Live action and animated.');
INSERT INTO content_rating (id, name, description) VALUES (4, 'R', 'Restricted; very violent or sexual in content.');
INSERT INTO content_rating (id, name, description) VALUES (5, 'NR', 'Not Rated; reserved for the dankest of may-mays, may be disturbing. Usually NSFW, will generally get you fired if you look at these at work.');