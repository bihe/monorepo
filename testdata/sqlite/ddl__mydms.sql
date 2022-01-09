CREATE TABLE "DOCUMENTS" (
	"id"	varchar(36) NOT NULL,
	"title"	varchar(255) NOT NULL,
	"filename"	varchar(255) NOT NULL,
	"alternativeid"	varchar(128),
	"previewlink"	varchar(128),
	"amount"	decimal(10 , 0),
	"created"	date NOT NULL,
	"modified"	date,
	"taglist"	text,
	"senderlist"	text,
	"invoicenumber"	varchar(128),
	PRIMARY KEY("id")
);

CREATE INDEX "IX_DOCUMENTS_PK" ON "DOCUMENTS" (
	"id"
);
