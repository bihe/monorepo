CREATE TABLE "BOOKMARKS" (
	"id"	varchar(255),
	"path"	varchar(255) NOT NULL,
	"display_name"	varchar(128) NOT NULL,
	"url"	varchar(512) NOT NULL,
	"sort_order"	integer NOT NULL DEFAULT 0,
	"type"	integer NOT NULL DEFAULT 0,
	"user_name"	varchar(128) NOT NULL,
	"created"	datetime NOT NULL,
	"modified"	datetime,
	"child_count"	integer NOT NULL DEFAULT 0,
	"access_count"	integer NOT NULL DEFAULT 0,
	"favicon"	varchar(128) NOT NULL,
	PRIMARY KEY("id")
);


CREATE INDEX "IX_PATH" ON "BOOKMARKS" (
	"path"
);

CREATE INDEX "IX_PATH_USER" ON "BOOKMARKS" (
	"path",
	"user_name"
);

CREATE INDEX "IX_SORT_ORDER" ON "BOOKMARKS" (
	"url"
);

CREATE INDEX "IX_USER" ON "BOOKMARKS" (
	"user_name"
);
