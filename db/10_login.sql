USE `login`;

LOCK TABLES `USERSITE` WRITE;
/*!40000 ALTER TABLE `USERSITE` DISABLE KEYS */;
INSERT INTO `USERSITE` VALUES ('bookmarks','henrik@binggl.net','http://localhost:3003','Admin;User','2020-01-03'),
('login','henrik@binggl.net','http://localhost:3001','Admin;User','2020-01-03'),
('mydms','henrik@binggl.net','http://localhost:3002','Admin;User','2020-01-03'),
('crypter','henrik@binggl.net','http://localhost:3004','Admin;User','2020-01-03'),
('onefrontend','henrik@binggl.net','http://localhost:3000','Admin;User','2020-03-08');

/*!40000 ALTER TABLE `USERSITE` ENABLE KEYS */;
UNLOCK TABLES;
