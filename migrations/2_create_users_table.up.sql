CREATE TABLE IF NOT EXISTS users (
	id VARCHAR(255) NOT NULL PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
	username VARCHAR(255) NULL,
	password VARCHAR(255) NULL
);