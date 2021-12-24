package common

var setupQueries = []string{
	`CREATE OR REPLACE TABLE emotes_to_roles (
	guild varchar(128) not null,
	emote varchar(128) not null,
	role bigint unsigned not null,
	PRIMARY KEY (guild, emote)
);`,
	`CREATE OR REPLACE TABLE role_react_messages (
	guild   varchar(128) not null,
	channel varchar(128) not null,
	message varchar(128) not null,
	maxpicks int unsigned,
	PRIMARY KEY (guild, channel, message)
);`,
	`CREATE OR REPLACE TABLE user_prefix (
	user varchar(128) primary key not null,
	prefix varchar(12) default '!' not null
);`,
}

//var setupQuery = `CREATE DATABASE hydaelyn;
