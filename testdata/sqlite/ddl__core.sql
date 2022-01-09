CREATE TABLE "USERSITE" (
	"name"	TEXT NOT NULL,
	"user"	TEXT NOT NULL,
	"url"	TEXT NOT NULL,
	"permission_list"	TEXT NOT NULL,
	"created"	DATE NOT NULL DEFAULT (datetime('now','localtime')),
	PRIMARY KEY("name","user")
);
