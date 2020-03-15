-- MySQL dump 10.16  Distrib 10.1.44-MariaDB, for debian-linux-gnu (x86_64)
--
-- Host: localhost    Database: bookmarks
-- ------------------------------------------------------
-- Server version	10.4.12-MariaDB-1:10.4.12+maria~bionic

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

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
  `access_count` int(11) NOT NULL DEFAULT 0,
  `favicon` varchar(128) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `IX_PATH` (`path`),
  KEY `IX_SORT_ORDER` (`sort_order`),
  KEY `IX_USER` (`user_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Current Database: `login`
--

CREATE DATABASE /*!32312 IF NOT EXISTS*/ `login` /*!40100 DEFAULT CHARACTER SET latin1 */;

USE `login`;

--
-- Table structure for table `LOGINS`
--

DROP TABLE IF EXISTS `LOGINS`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `LOGINS` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user` varchar(128) NOT NULL,
  `created` date NOT NULL DEFAULT current_timestamp(),
  `type` tinyint(4) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `user` (`user`)
) ENGINE=InnoDB AUTO_INCREMENT=224 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `USERSITE`
--

DROP TABLE IF EXISTS `USERSITE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `USERSITE` (
  `name` varchar(128) NOT NULL,
  `user` varchar(128) NOT NULL,
  `url` varchar(256) NOT NULL,
  `permission_list` varchar(256) NOT NULL,
  `created` date NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`name`,`user`)
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

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2020-03-15 18:34:22
