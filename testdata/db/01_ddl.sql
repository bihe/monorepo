--
-- Current Database: `bookmarks`
--

CREATE DATABASE /*!32312 IF NOT EXISTS*/ `bookmarks` /*!40100 DEFAULT CHARACTER SET latin1 */;

USE `bookmarks`;

--
-- Table structure for table `BOOKMARKS`
--

DROP TABLE IF EXISTS `BOOKMARKS`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `BOOKMARKS` (
  `id` varchar(255) CHARACTER SET utf8mb4 NOT NULL,
  `path` varchar(255) CHARACTER SET utf8mb4 NOT NULL,
  `display_name` varchar(128) CHARACTER SET utf8mb4 NOT NULL,
  `url` varchar(512) CHARACTER SET utf8mb4 NOT NULL,
  `sort_order` int(11) NOT NULL DEFAULT 0,
  `type` int(11) NOT NULL,
  `user_name` varchar(128) CHARACTER SET utf8mb4 NOT NULL,
  `created` datetime(6) NOT NULL,
  `modified` datetime(6) DEFAULT NULL,
  `child_count` int(11) NOT NULL DEFAULT 0,
  `highlight` int(11) NOT NULL DEFAULT 0,
  `favicon` varchar(128) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `IX_PATH` (`path`),
  KEY `IX_SORT_ORDER` (`sort_order`),
  KEY `IX_USER` (`user_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Current Database: `mydms`
--

CREATE DATABASE /*!32312 IF NOT EXISTS*/ `mydms` /*!40100 DEFAULT CHARACTER SET utf8 */;

USE `mydms`;

--
-- Table structure for table `DOCUMENTS`
--

DROP TABLE IF EXISTS `DOCUMENTS`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `DOCUMENTS` (
  `id` char(36) NOT NULL,
  `title` varchar(255) NOT NULL,
  `filename` varchar(255) NOT NULL,
  `alternativeid` varchar(128) DEFAULT NULL,
  `previewlink` varchar(128) DEFAULT NULL,
  `amount` decimal(10,0) DEFAULT NULL,
  `created` datetime NOT NULL DEFAULT current_timestamp(),
  `modified` datetime DEFAULT NULL,
  `taglist` text DEFAULT NULL,
  `senderlist` text DEFAULT NULL,
  `invoicenumber` varchar(128) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `alternativeid_unique` (`alternativeid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `DOCUMENTS_TO_SENDERS`
--

DROP TABLE IF EXISTS `DOCUMENTS_TO_SENDERS`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `DOCUMENTS_TO_SENDERS` (
  `document_id` char(36) NOT NULL,
  `sender_id` bigint(20) NOT NULL,
  PRIMARY KEY (`document_id`,`sender_id`),
  KEY `fk_sender_document_id` (`sender_id`),
  CONSTRAINT `fk_document_sender_id` FOREIGN KEY (`document_id`) REFERENCES `DOCUMENTS` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_sender_document_id` FOREIGN KEY (`sender_id`) REFERENCES `SENDERS` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `DOCUMENTS_TO_TAGS`
--

DROP TABLE IF EXISTS `DOCUMENTS_TO_TAGS`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `DOCUMENTS_TO_TAGS` (
  `document_id` char(36) NOT NULL,
  `tag_id` bigint(20) NOT NULL,
  PRIMARY KEY (`document_id`,`tag_id`),
  KEY `fk_tag_document_id` (`tag_id`),
  CONSTRAINT `fk_document_tag_id` FOREIGN KEY (`document_id`) REFERENCES `DOCUMENTS` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_tag_document_id` FOREIGN KEY (`tag_id`) REFERENCES `TAGS` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `SENDERS`
--

DROP TABLE IF EXISTS `SENDERS`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `SENDERS` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `sender_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=424 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `TAGS`
--

DROP TABLE IF EXISTS `TAGS`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `TAGS` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `tag_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=244 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `UPLOADS`
--

DROP TABLE IF EXISTS `UPLOADS`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `UPLOADS` (
  `id` char(36) NOT NULL,
  `filename` varchar(255) NOT NULL,
  `mimetype` varchar(255) NOT NULL,
  `created` datetime NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

