CREATE TABLE "DOCUMENTS" (
  "id" TEXT NOT NULL,
  "title" TEXT NOT NULL,
  "filename" TEXT NOT NULL,
  "alternativeid" TEXT NOT NULL UNIQUE,
  "previewlink" TEXT,
  "amount" NUMERIC,
  "created" datetime NOT NULL,
  "modified" datetime,
  "taglist" TEXT,
  "senderlist" TEXT,
  "invoicenumber" TEXT,
  PRIMARY KEY("id")
);

CREATE TABLE "SENDERS" (
  "id" INTEGER NOT NULL,
  "name" TEXT NOT NULL UNIQUE,
  PRIMARY KEY ("id"  AUTOINCREMENT)
);

CREATE TABLE "TAGS" (
  "id" INTEGER NOT NULL,
  "name" TEXT NOT NULL UNIQUE,
  PRIMARY KEY ("id"  AUTOINCREMENT)
);

CREATE TABLE "DOCUMENTS_TO_SENDERS" (
  "document_id" TEXT NOT NULL,
  "sender_id" INTEGER NOT NULL,
  PRIMARY KEY ("document_id","sender_id"),
  CONSTRAINT "fk_document_sender_id" FOREIGN KEY ("document_id") REFERENCES "DOCUMENTS" ("id") ON DELETE CASCADE,
  CONSTRAINT "fk_sender_document_id" FOREIGN KEY ("sender_id") REFERENCES "SENDERS" ("id") ON DELETE CASCADE
);

CREATE TABLE "DOCUMENTS_TO_TAGS" (
  "document_id" TEXT NOT NULL,
  "tag_id" INTEGER NOT NULL,
  PRIMARY KEY ("document_id","tag_id"),
  CONSTRAINT "fk_document_tag_id" FOREIGN KEY ("document_id") REFERENCES "DOCUMENTS" ("id") ON DELETE CASCADE,
  CONSTRAINT "fk_tag_document_id" FOREIGN KEY ("tag_id") REFERENCES "TAGS" ("id") ON DELETE CASCADE
);

