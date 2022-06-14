/*
 * This file is part of the RonDB REST API Server
 * Copyright (c) 2022 Hopsworks AB
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, version 3.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package common

import (
	"fmt"
	"strconv"
	"strings"
)

var databases map[string][][]string = make(map[string][][]string)

func init() {
	db := "bench"
	databases[db] = [][]string{
		benchmarSchema(db, 1000),

		{ // clean up commands
			"DROP DATABASE " + db,
		},
	}

	db = "DB000"
	databases[db] = [][]string{
		{
			// setup commands
			"DROP DATABASE IF EXISTS " + db,
			"CREATE DATABASE " + db,
			"USE " + db,
		},

		{ // clean up commands
			"DROP DATABASE " + db,
		},
	}

	db = "DB001"
	databases[db] = [][]string{
		{
			// setup commands
			"DROP DATABASE IF EXISTS " + db,
			"CREATE DATABASE " + db,
			"USE " + db,

			"CREATE TABLE table_1(id0 VARCHAR(10), col_0 VARCHAR(100), col_1 VARCHAR(100), col_2 VARCHAR(100), PRIMARY KEY(id0))",
			"INSERT INTO table_1 VALUES('id0_data', 'col_0_data', 'col_1_data', 'col_2_data')",
		},

		{ // clean up commands
			"DROP DATABASE " + db,
		},
	}

	db = "DB002"
	databases[db] = [][]string{
		{
			// setup commands
			"DROP DATABASE IF EXISTS " + db,
			"CREATE DATABASE " + db,
			"USE " + db,

			"CREATE TABLE table_1(id0 VARCHAR(10), id1 VARCHAR(10), col_0 VARCHAR(100), col_1 VARCHAR(100), col_2 VARCHAR(100), PRIMARY KEY(id0, id1))",
			"INSERT INTO table_1 VALUES('id0_data', 'id1_data', 'col_0_data', 'col_1_data', 'col_2_data')",
		},

		{ // clean up commands
			"DROP DATABASE " + db,
		},
	}

	db = "DB003"
	databases[db] = [][]string{
		{
			// setup commands
			"DROP DATABASE IF EXISTS " + db,
			"CREATE DATABASE " + db,
			"USE " + db,

			"CREATE TABLE `date_table` ( `id0` int NOT NULL, `col0` date DEFAULT NULL, `col1` time DEFAULT NULL, `col2` datetime DEFAULT NULL, `col3` timestamp NULL DEFAULT NULL, `col4` year DEFAULT NULL, PRIMARY KEY (`id0`))",
			"insert into date_table values(1, \"1111-11-11\", \"11:11:11\", \"1111-11-11 11:11:11\", \"1970-11-11 11:11:11\", \"11\")",
			"insert into date_table set id0=2",

			"CREATE TABLE `arrays_table` ( `id0` int NOT NULL, `col0` char(100) DEFAULT NULL, `col2` varchar(100) DEFAULT NULL, `col3` binary(100) DEFAULT NULL, `col4` varbinary(100)      DEFAULT NULL, PRIMARY KEY (`id0`))",
			"insert into arrays_table values (1, \"abcd\", \"abcd\", 0xFFFF, 0xFFFF)",
			"insert into arrays_table set id0=2",

			"CREATE TABLE `set_table` ( `id0` int NOT NULL, `col0` enum('a','b','c','d') DEFAULT NULL, `col1` set('a','b','c','d') DEFAULT NULL, PRIMARY KEY (`id0`))",
			"INSERT INTO `set_table` VALUES (1,'a','a')",
			"INSERT INTO `set_table` VALUES (2,'b','a,b')",
			"insert into set_table set id0=3",

			"CREATE TABLE `special_table` ( `id0` int NOT NULL, `col0` geometry DEFAULT NULL, `col1` point DEFAULT NULL, `col2` linestring DEFAULT NULL, `col3` polygon DEFAULT NULL,       `col4` geomcollection DEFAULT NULL, `col5` multilinestring DEFAULT NULL, `col6` multipoint DEFAULT NULL, `col7` multipolygon DEFAULT NULL, PRIMARY KEY (`id0`))",
			"insert into special_table set id0=1, col0=ST_GeomFromText('POINT(1 1)'), col1=ST_GeomFromText('POINT(1 1)'), col2=ST_GeomFromText('LineString(1 1,2 2,3 3)'), col3=ST_GeomFromText('Polygon((0 0,0 3,3 0,0 0),(1 1,1 2,2 1,1 1))'), col7=ST_GeomFromText('MultiPolygon(((0 0,0 3,3 3,3 0,0 0),(1 1,1 2,2 2,2 1,1 1)))'),col4=ST_GeomFromText('GeometryCollection(Point(1 1),LineString(2 2, 3 3))'),col6=ST_MPointFromText('MULTIPOINT (1 1, 2 2, 3 3)'),col5=ST_GeomFromText('MultiLineString((1 1,2 2,3 3),(4 4,5 5))')",
			"insert into special_table set id0=2",

			"CREATE TABLE `number_table` ( `id0` int NOT NULL, `col0` tinyint DEFAULT NULL, `col1` smallint DEFAULT NULL, `col2` mediumint DEFAULT NULL, `col3` int DEFAULT NULL, `col4` bigint DEFAULT NULL, `col5` decimal(10, 0) DEFAULT NULL, `col6` float DEFAULT NULL, `col7` double DEFAULT NULL, `col8` bit(1) DEFAULT NULL, PRIMARY KEY (`id0`))",
			"INSERT INTO `number_table` VALUES (1,99,99,99,99,99,99,99.99,99.99,true)",
			"insert into number_table set id0=2",

			"CREATE TABLE `blob_table` ( `id0` int NOT NULL, `col0` tinyblob, `col1` blob, `col2` mediumblob, `col3` longblob, `col4` tinytext, `col5` mediumtext, `col6` longtext, PRIMARY KEY (`id0`))",
			"insert into blob_table values(1, 0xFFFF, 0xFFFF, 0xFFFF,  0xFFFF, \"abcd\", \"abcd\", \"abcd\")",
			"insert into blob_table set id0=2",
		},

		{ // clean up commands
			"DROP DATABASE " + db,
		},
	}

	// signed and unsigned number data types
	db = "DB004"
	databases[db] = [][]string{
		{
			// setup commands
			"DROP DATABASE IF EXISTS " + db,
			"CREATE DATABASE " + db,
			"USE " + db,

			"CREATE TABLE int_table(id0 INT, id1 INT UNSIGNED, col0 INT, col1 INT UNSIGNED, PRIMARY KEY(id0, id1))",
			"INSERT INTO  int_table VALUES(2147483647,4294967295,2147483647,4294967295)",
			"INSERT INTO  int_table VALUES(-2147483648,0,-2147483648,0)",
			"INSERT INTO  int_table VALUES(0,0,0,0)",
			"INSERT INTO  int_table set id0=1, id1=1", // NULL values for non primary columns

			// this table only has primary keys
			"CREATE TABLE int_table1(id0 INT, id1 INT UNSIGNED, PRIMARY KEY(id0, id1))",
			"INSERT INTO  int_table1 VALUES(0,0)",
		},

		{ // clean up commands
			"DROP DATABASE " + db,
		},
	}

	db = "DB005"
	databases[db] = [][]string{
		{
			// setup commands
			"DROP DATABASE IF EXISTS " + db,
			"CREATE DATABASE " + db,
			"USE " + db,

			"CREATE TABLE bigint_table(id0 BIGINT, id1 BIGINT UNSIGNED, col0 BIGINT, col1 BIGINT UNSIGNED, PRIMARY KEY(id0, id1))",
			"INSERT INTO  bigint_table VALUES(9223372036854775807,18446744073709551615,9223372036854775807,18446744073709551615)",
			"INSERT INTO  bigint_table VALUES(-9223372036854775808,0,-9223372036854775808,0)",
			"INSERT INTO  bigint_table VALUES(0,0,0,0)",
			"INSERT INTO  bigint_table set id0=1, id1=1", // NULL values for non primary columns
		},

		{ // clean up commands
			"DROP DATABASE " + db,
		},
	}

	db = "DB006"
	databases[db] = [][]string{
		{
			// setup commands
			"DROP DATABASE IF EXISTS " + db,
			"CREATE DATABASE " + db,
			"USE " + db,

			"CREATE TABLE tinyint_table(id0 TINYINT, id1 TINYINT UNSIGNED, col0 TINYINT, col1 TINYINT UNSIGNED, PRIMARY KEY(id0, id1))",
			"INSERT INTO  tinyint_table VALUES(127,255,127,255)",
			"INSERT INTO  tinyint_table VALUES(-128,0,-128,0)",
			"INSERT INTO  tinyint_table VALUES(0,0,0,0)",
			"INSERT INTO  tinyint_table set id0=1, id1=1", // NULL values for non primary columns
		},

		{ // clean up commands
			"DROP DATABASE " + db,
		},
	}

	db = "DB007"
	databases[db] = [][]string{
		{
			// setup commands
			"DROP DATABASE IF EXISTS " + db,
			"CREATE DATABASE " + db,
			"USE " + db,

			"CREATE TABLE smallint_table(id0 SMALLINT, id1 SMALLINT UNSIGNED, col0 SMALLINT, col1 SMALLINT UNSIGNED, PRIMARY KEY(id0, id1))",
			"INSERT INTO  smallint_table VALUES(32767,65535,32767,65535)",
			"INSERT INTO  smallint_table VALUES(-32768,0,-32768,0)",
			"INSERT INTO  smallint_table VALUES(0,0,0,0)",
			"INSERT INTO  smallint_table set id0=1, id1=1", // NULL values for non primary columns

		},

		{ // clean up commands
			"DROP DATABASE " + db,
		},
	}

	db = "DB008"
	databases[db] = [][]string{
		{
			// setup commands
			"DROP DATABASE IF EXISTS " + db,
			"CREATE DATABASE " + db,
			"USE " + db,

			"CREATE TABLE mediumint_table(id0 MEDIUMINT, id1 MEDIUMINT UNSIGNED, col0 MEDIUMINT, col1 MEDIUMINT UNSIGNED, PRIMARY KEY(id0, id1))",
			"INSERT INTO  mediumint_table VALUES(8388607,16777215,8388607,16777215)",
			"INSERT INTO  mediumint_table VALUES(-8388608,0,-8388608,0)",
			"INSERT INTO  mediumint_table VALUES(0,0,0,0)",
			"INSERT INTO  mediumint_table set id0=1, id1=1", // NULL values for non primary columns

		},

		{ // clean up commands
			"DROP DATABASE " + db,
		},
	}

	db = "DB009"
	databases[db] = [][]string{
		{
			// setup commands
			"DROP DATABASE IF EXISTS " + db,
			"CREATE DATABASE " + db,
			"USE " + db,

			"CREATE TABLE float_table1(id0 INT, col0 FLOAT, col1 FLOAT UNSIGNED, PRIMARY KEY(id0))",
			"INSERT INTO  float_table1 VALUES(1,-123.123,123.123)",
			"INSERT INTO  float_table1 VALUES(0,0,0)",
			"INSERT INTO  float_table1 set id0=2", // NULL values for non primary columns

			"CREATE TABLE float_table2(id0 FLOAT, col0 FLOAT, col1 FLOAT UNSIGNED, PRIMARY KEY(id0))",
		},

		{ // clean up commands
			"DROP DATABASE " + db,
		},
	}

	db = "DB010"
	databases[db] = [][]string{
		{
			// setup commands
			"DROP DATABASE IF EXISTS " + db,
			"CREATE DATABASE " + db,
			"USE " + db,

			"CREATE TABLE double_table1(id0 INT, col0 DOUBLE, col1 DOUBLE UNSIGNED, PRIMARY KEY(id0))",
			"INSERT INTO  double_table1 VALUES(1,-123.123,123.123)",
			"INSERT INTO  double_table1 VALUES(0,0,0)",
			"INSERT INTO  double_table1 set id0=2", // NULL values for non primary columns

			"CREATE TABLE double_table2(id0 DOUBLE, col0 DOUBLE, col1 DOUBLE UNSIGNED, PRIMARY KEY(id0))",
		},

		{ // clean up commands
			"DROP DATABASE " + db,
		},
	}

	db = "DB011"
	databases[db] = [][]string{
		{
			// setup commands
			"DROP DATABASE IF EXISTS " + db,
			"CREATE DATABASE " + db,
			"USE " + db,

			"CREATE TABLE decimal_table(id0 DECIMAL(10,5), id1 DECIMAL(10,5) UNSIGNED, col0 DECIMAL(10,5), col1 DECIMAL(10,5) UNSIGNED, PRIMARY KEY(id0, id1))",
			"INSERT INTO  decimal_table VALUES(-12345.12345,12345.12345,-12345.12345,12345.12345)",
			"INSERT INTO  decimal_table set id0=-67890.12345, id1=67890.12345",
		},

		{ // clean up commands
			"DROP DATABASE " + db,
		},
	}

	db = "DB012"
	databases[db] = SchemaTextualColumns("char", db, 100)

	db = "DB013"
	databases[db] = [][]string{
		{
			// setup commands
			"DROP DATABASE IF EXISTS " + db,
			"CREATE DATABASE " + db,
			"USE " + db,

			// blobs in PK is not supported by RonDB
			"CREATE TABLE blob_table(id0 int, col0 blob, col1 int,  PRIMARY KEY(id0))",
			"INSERT INTO  blob_table VALUES(1,0xFFFF, 1)",
			"CREATE TABLE text_table(id0 int, col0 text, col1 int, PRIMARY KEY(id0))",
			"INSERT INTO  text_table VALUES(1,\"FFFF\", 1)",
		},

		{ // clean up commands
			"DROP DATABASE " + db,
		},
	}

	db = "DB014" //varchar
	databases[db] = SchemaTextualColumns("VARCHAR", db, 50)

	db = "DB015" //long varchar
	databases[db] = SchemaTextualColumns("VARCHAR", db, 256)

	db = "DB016" //binary fix size
	databases[db] = SchemaTextualColumns("BINARY", db, 100)

	db = "DB017" //varbinary
	databases[db] = SchemaTextualColumns("VARBINARY", db, 100)

	db = "DB018" //long varbinary
	databases[db] = SchemaTextualColumns("VARBINARY", db, 256)

	db = "DB019"
	databases[db] = [][]string{
		{
			// setup commands
			"DROP DATABASE IF EXISTS " + db,
			"CREATE DATABASE " + db,
			"USE " + db,

			// blobs in PK is not supported by RonDB
			"CREATE TABLE `date_table` ( `id0`  date, `col0` date DEFAULT NULL, PRIMARY KEY (`id0`))",
			"insert into date_table values( \"1111-11-11\", \"1111:11:11\")",
			"insert into date_table set id0= \"1111-11-12\" ",
		},

		{ // clean up commands
			"DROP DATABASE " + db,
		},
	}

	db = "DB020"
	databases[db] = [][]string{
		{
			// setup commands
			"DROP DATABASE IF EXISTS " + db,
			"CREATE DATABASE " + db,
			"USE " + db,

			// blobs in PK is not supported by RonDB
			"CREATE TABLE `date_table0` ( `id0`  datetime(0), `col0` datetime(0) DEFAULT NULL, PRIMARY KEY (`id0`))",
			"insert into date_table0 values( \"1111-11-11 11:11:11\", \"1111-11-11 11:11:11\")",
			"insert into date_table0 set id0= \"1111-11-12 11:11:11\"",

			"CREATE TABLE `date_table3` ( `id0`  datetime(3), `col0` datetime(3) DEFAULT NULL, PRIMARY KEY (`id0`))",
			"insert into date_table3 values( \"1111-11-11 11:11:11.123\", \"1111-11-11 11:11:11.123\")",
			"insert into date_table3 set id0= \"1111-11-12 11:11:11.123\"",

			"CREATE TABLE `date_table6` ( `id0`  datetime(6), `col0` datetime(6) DEFAULT NULL, PRIMARY KEY (`id0`))",
			"insert into date_table6 values( \"1111-11-11 11:11:11.123456\", \"1111-11-11 11:11:11.123456\")",
			"insert into date_table6 set id0= \"1111-11-12 11:11:11.123456\"",
		},

		{ // clean up commands
			"DROP DATABASE " + db,
		},
	}

	db = "DB021"
	databases[db] = [][]string{
		{
			// setup commands
			"DROP DATABASE IF EXISTS " + db,
			"CREATE DATABASE " + db,
			"USE " + db,

			// blobs in PK is not supported by RonDB
			"CREATE TABLE `time_table0` ( `id0`  time(0), `col0` time(0) DEFAULT NULL, PRIMARY KEY (`id0`))",
			"insert into time_table0 values( \"11:11:11\", \"11:11:11\")",
			"insert into time_table0 set id0= \"12:11:11\"",

			"CREATE TABLE `time_table3` ( `id0`  time(3), `col0` time(3) DEFAULT NULL, PRIMARY KEY (`id0`))",
			"insert into time_table3 values( \"11:11:11.123\", \"11:11:11.123\")",
			"insert into time_table3 set id0= \"12:11:11.123\"",

			"CREATE TABLE `time_table6` ( `id0` time(6), `col0` time(6) DEFAULT NULL, PRIMARY KEY (`id0`))",
			"insert into time_table6 values( \"11:11:11.123456\", \"11:11:11.123456\")",
			"insert into time_table6 set id0= \"12:11:11.123456\"",
		},

		{ // clean up commands
			"DROP DATABASE " + db,
		},
	}

	db = "DB022"
	databases[db] = [][]string{
		{
			// setup commands
			"DROP DATABASE IF EXISTS " + db,
			"CREATE DATABASE " + db,
			"USE " + db,

			// blobs in PK is not supported by RonDB
			"CREATE TABLE `ts_table0` ( `id0`  timestamp(0), `col0` timestamp(0) DEFAULT NULL, PRIMARY KEY (`id0`))",
			"insert into ts_table0 values( \"2022-11-11 11:11:11\", \"2022-11-11 11:11:11\")",
			"insert into ts_table0 set id0= \"2022-11-12 11:11:11\"",

			"CREATE TABLE `ts_table3` ( `id0`  timestamp(3), `col0` timestamp(3) DEFAULT NULL, PRIMARY KEY (`id0`))",
			"insert into ts_table3 values( \"2022-11-11 11:11:11.123\", \"2022-11-11 11:11:11.123\")",
			"insert into ts_table3 set id0= \"2022-11-12 11:11:11.123\"",

			"CREATE TABLE `ts_table6` ( `id0`  timestamp(6), `col0` timestamp(6) DEFAULT NULL, PRIMARY KEY (`id0`))",
			"insert into ts_table6 values( \"2022-11-11 11:11:11.123456\", \"2022-11-11 11:11:11.123456\")",
			"insert into ts_table6 set id0= \"2022-11-12 11:11:11.123456\"",
		},

		{ // clean up commands
			"DROP DATABASE " + db,
		},
	}

	db = "DB023"
	databases[db] = [][]string{
		{
			// setup commands
			"DROP DATABASE IF EXISTS " + db,
			"CREATE DATABASE " + db,
			"USE " + db,

			// blobs in PK is not supported by RonDB
			"CREATE TABLE `year_table` ( `id0`  year, `col0` year DEFAULT NULL, PRIMARY KEY (`id0`))",
			"insert into year_table values( \"2022\", \"2022\")",
			"insert into year_table set id0=\"2023\"",
		},

		{ // clean up commands
			"DROP DATABASE " + db,
		},
	}

	db = "DB024"
	databases[db] = [][]string{
		{
			// setup commands
			"DROP DATABASE IF EXISTS " + db,
			"CREATE DATABASE " + db,
			"USE " + db,

			// blobs in PK is not supported by RonDB
			"CREATE TABLE `bit_table` ( `id0`  binary(100), `col0` bit(1) DEFAULT NULL, `col1` bit(3) DEFAULT NULL, `col2` bit(25) DEFAULT NULL,`col3` bit(39) DEFAULT NULL, col4 bit(64) DEFAULT NULL, PRIMARY KEY (`id0`))",
			"insert into bit_table values(1,  b'1',  b'111', b'1111111111111111111111111', b'111111111111111111111111111111111111111', b'1111111111111111111111111111111111111111111111111111111111111111')",
			"insert into bit_table values(2,  b'0',  b'000', b'0000000000000000000000000', b'000000000000000000000000000000000000000', b'0000000000000000000000000000000000000000000000000000000000000000')",
			"insert into bit_table set id0=\"3\"",
		},

		{ // clean up commands
			"DROP DATABASE " + db,
		},
	}

	db = "hopsworks"
	databases[db] = [][]string{
		{
			// setup commands
			"DROP DATABASE IF EXISTS " + db,
			"CREATE DATABASE " + db,
			"USE " + db,

			"CREATE TABLE `users` (" +
				"`uid` int NOT NULL AUTO_INCREMENT," +
				"`username` varchar(10) CHARACTER SET latin1 COLLATE latin1_general_cs NOT NULL," +
				"`password` varchar(128) CHARACTER SET latin1 COLLATE latin1_general_cs NOT NULL," +
				"`email` varchar(150) CHARACTER SET latin1 COLLATE latin1_general_cs DEFAULT NULL," +
				"`fname` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL," +
				"`lname` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL," +
				"`activated` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP," +
				"`title` varchar(10) CHARACTER SET latin1 COLLATE latin1_general_cs DEFAULT '-'," +
				"`false_login` int NOT NULL DEFAULT '-1'," +
				"`status` int NOT NULL DEFAULT '-1'," +
				"`isonline` int NOT NULL DEFAULT '-1'," +
				"`secret` varchar(20) CHARACTER SET latin1 COLLATE latin1_general_cs DEFAULT NULL," +
				"`validation_key` varchar(128) CHARACTER SET latin1 COLLATE latin1_general_cs DEFAULT NULL," +
				"`validation_key_updated` timestamp NULL DEFAULT NULL," +
				"`validation_key_type` varchar(20) COLLATE latin1_general_cs DEFAULT NULL," +
				"`mode` int NOT NULL DEFAULT '0'," +
				"`password_changed` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP," +
				"`notes` varchar(500) CHARACTER SET latin1 COLLATE latin1_general_cs DEFAULT '-'," +
				"`max_num_projects` int NOT NULL," +
				"`num_active_projects` int NOT NULL DEFAULT '0'," +
				"`num_created_projects` int NOT NULL DEFAULT '0'," +
				"`two_factor` tinyint(1) NOT NULL DEFAULT '1'," +
				"`tours_state` tinyint(1) NOT NULL DEFAULT '0'," +
				"`salt` varchar(128) CHARACTER SET latin1 COLLATE latin1_general_cs NOT NULL DEFAULT ''," +
				"PRIMARY KEY (`uid`)," +
				"UNIQUE KEY `username` (`username`)," +
				"UNIQUE KEY `email` (`email`))",

			"CREATE TABLE `project` (" +
				"`id` int NOT NULL AUTO_INCREMENT," +
				"`inode_pid` bigint NOT NULL," +
				"`inode_name` varchar(255) CHARACTER SET latin1 COLLATE latin1_general_cs NOT NULL," +
				"`partition_id` bigint NOT NULL," +
				"`projectname` varchar(100) CHARACTER SET latin1 COLLATE latin1_general_cs NOT NULL," +
				"`username` varchar(150) CHARACTER SET latin1 COLLATE latin1_general_cs NOT NULL," +
				"`created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP," +
				"`retention_period` date DEFAULT NULL," +
				"`archived` tinyint(1) DEFAULT '0'," +
				"`logs` tinyint(1) DEFAULT '0'," +
				"`deleted` tinyint(1) DEFAULT '0'," +
				"`description` varchar(2000) CHARACTER SET latin1 COLLATE latin1_general_cs DEFAULT NULL," +
				"`payment_type` varchar(255) CHARACTER SET latin1 COLLATE latin1_general_cs NOT NULL DEFAULT 'PREPAID'," +
				"`last_quota_update` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP," +
				"`kafka_max_num_topics` int NOT NULL DEFAULT '100'," +
				"`docker_image` varchar(255) CHARACTER SET latin1 COLLATE latin1_general_cs DEFAULT NULL," +
				"`python_env_id` int DEFAULT NULL," +
				"PRIMARY KEY (`id`)," +
				"UNIQUE KEY `projectname` (`projectname`)," +
				"UNIQUE KEY `inode_pid` (`inode_pid`,`inode_name`,`partition_id`)," +
				"KEY `user_idx` (`username`)," +
				// "CONSTRAINT `FK_149_289` FOREIGN KEY (`inode_pid`, `inode_name`, `partition_id`) REFERENCES `hops`.`hdfs_inodes` (`parent_id`, `name`, `partition_id`) ON DELETE CASCADE," +
				"CONSTRAINT `FK_262_290` FOREIGN KEY (`username`) REFERENCES `users` (`email`))",

			"CREATE TABLE `project_team` (" +
				"`project_id` int NOT NULL," +
				"`team_member` varchar(150) CHARACTER SET latin1 COLLATE latin1_general_cs NOT NULL," +
				"`team_role` varchar(32) CHARACTER SET latin1 COLLATE latin1_general_cs NOT NULL," +
				"`added` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP," +
				"PRIMARY KEY (`project_id`,`team_member`)," +
				"KEY `team_member` (`team_member`)," +
				"CONSTRAINT `FK_262_304` FOREIGN KEY (`team_member`) REFERENCES `users` (`email`) ON DELETE CASCADE," +
				"CONSTRAINT `FK_284_303` FOREIGN KEY (`project_id`) REFERENCES `project` (`id`) ON DELETE CASCADE )",

			"CREATE TABLE `api_key` (" +
				"`id` int NOT NULL AUTO_INCREMENT," +
				"`prefix` varchar(45) COLLATE latin1_general_cs NOT NULL," +
				"`secret` varchar(512) COLLATE latin1_general_cs NOT NULL," +
				"`salt` varchar(256) COLLATE latin1_general_cs NOT NULL," +
				"`created` timestamp NOT NULL," +
				"`modified` timestamp NOT NULL," +
				"`name` varchar(45) COLLATE latin1_general_cs NOT NULL," +
				"`user_id` int NOT NULL," +
				"`reserved` tinyint(1) DEFAULT '0'," +
				"PRIMARY KEY (`id`)," +
				"UNIQUE KEY `prefix_UNIQUE` (`prefix`)," +
				"UNIQUE KEY `index4` (`user_id`,`name`)," +
				"KEY `fk_api_key_1_idx` (`user_id`)," +
				"CONSTRAINT `fk_api_key_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`uid`) ON DELETE CASCADE)",

			"INSERT INTO `users` VALUES (10000,'meb10000','12fa520ec8f65d3a6feacfa97a705e622e1fea95b80b521ec016e43874dfed5a','admin@hopsworks.ai','Admin','Admin','2015-05-15 10:22:36','Mr',0,2,1,'V3WBPS4G2WMQ53VA',NULL,NULL,NULL,0,'2015-04-28 15:18:42',NULL,30,2,1,0,3,'+9mTLmYSpnZROFEJEaednw8+GDH/s2J1QuRZy8okxW5myI/q8ek8Xu+ab5CyE9GzhWX6Sa4cr7KX8cAHi5IC4g==');",
			"INSERT INTO `users` VALUES (10179,'onlinefs','0563740c85f6835a5140e29ecb38a8cda9619af28c5b5bebdb17d6bc39823bc8','onlinefs@hopsworks.ai','OnlineFS','Server','2021-03-16 16:17:00','Mr',-1,2,0,'V3WBPS4G2WMQ53VA',NULL,NULL,NULL,0,'2021-03-16 16:17:00',NULL,30,0,0,0,3,'PRkXn867Zzj22tocboz4kvvMtbOmjIdMWZJ++m0KSGqX0bPJRAD87zdzwPs+TRyVIjfI/PSP+5M9b8s98JVkxQ==');",
			"INSERT INTO `users` VALUES (10178,'srvmanager','kkiulkflnccswpvtndwcjmbbibmpwdfywwuxgbqymwogenhcmnoqoyuzkzgeyrapmwsvemksckvbgmliyrzfzoudarwiopjvsgoqfzmbfkxyhgkshjgkfeiphmxpozgl','serving@hopsworks.se','Serving','Manager','2015-05-15 10:22:36','Mr',-1,2,0,'V3WBPS4G2WMQ53VA',NULL,NULL,NULL,0,'2015-04-28 15:18:42',NULL,30,0,0,0,3,'nbyrnglnpfxhltlfrvcfsoyspwtzyonequiggafnvhpsbnftqzqovlquxthemegjrkkmhhvhmnefjuzgfwgducdnyulsbbkrjaotwdzbpzokqutbpqakdqrfqpxrxocg');",
			"INSERT INTO `users` VALUES (10180,'756162f8','b5df0121865326f5fce61e01436cf3b9f9fbc60011612532647855d77b2293f2','756162f80f2db313176ef1731b56d63fbbacfa8d@email.com','name','last','2022-06-01 13:10:07','-',0,2,1,'ECIEOTPGZFRENJVG',NULL,NULL,NULL,0,'2022-06-01 13:10:05',NULL,10,2,0,0,0,'3AGoW1OefH7AScVGwJdWCJPEuXgidnrnJ+0nwQfymGYYOjYQqgBWmJaqQVZpcVlRv7TASgt8K4bwJI+UYrbtEQ==');",
			"INSERT INTO `users` VALUES (10182,'2ac83d84','91273e97cafefcd93af546c503a30fe799725ff3a10fa93720aae0d02aae7f87','2ac83d840e694df7fc43325427c552122a8e69f9@email.com','name','last','2022-06-01 13:19:51','-',0,2,0,'FK5TX324MQZGJCVK',NULL,NULL,NULL,0,'2022-06-01 13:19:51',NULL,10,0,0,0,0,'5s/A8JkKZ7n22FLRNAULZUp9Wj4vefXwGcWQV0Zui+f+UNcMScOYxxE0N2utoXlBzq3ZggZQx5L/3BrDiTInVQ==');",
			"INSERT INTO `users` VALUES (10183,'c92f2da1','d2b9e3355eaba21973fb07c006b0530ba213ec36166dd43413908db2c7a90288','c92f2da1711e5da4b94c150d3174fa53a9bc91d7@email.com','name','last','2022-06-01 13:19:52','-',0,2,1,'4WKUUI3UYZAU5UKE',NULL,NULL,NULL,0,'2022-06-01 13:19:52',NULL,10,2,0,0,0,'G77uyxVvQs1PMxOhPuKqQKIVXxTFo3o63XGh6/SOp8LqTj0hRcht2qJpaEUrke6Vh5Sfp1u0WMTPsNcWisNshA==');",
			"INSERT INTO `users` VALUES (10185,'28c83a26','b62abaaedbcf61d5b6851b200641a140cfcf725a2086ceadf8444476f2e2315b','28c83a26b5906e88b146d75cfff5a2f988ea04bd@email.com','name','last','2022-06-01 13:20:27','-',0,2,0,'AZPVUBTVWE4XZUSH',NULL,NULL,NULL,0,'2022-06-01 13:20:26',NULL,10,0,0,0,0,'A0FPSIz7xbXFu717g17Lziu5ucHiRcov6UqtFD243UIAwDgltbtaP4IrMrVg3BtmFgT6Ge7YyVVm/9k43A1cWQ==');",
			"INSERT INTO `users` VALUES (10188,'e33a5381','c343fd88d1e0e17111a3fd65f353ac2a7702f348a0f22b1f43fba43a70826a23','e33a538197e34b92b6e97b2342a09d8576a00eed@email.com','name','last','2022-06-01 13:22:05','-',0,2,1,'VB7FD2XUWKBODIS3',NULL,NULL,NULL,0,'2022-06-01 13:22:05',NULL,10,2,0,0,0,'GHjmOwUJw2WnhvbFS1jeW27ee4Sd4niVs/JN5dK7L1Kn7H6mrSl3EpQAm+TTH10jmqnNBnG3PKhs2CoepI8gWg==');",
			"INSERT INTO `users` VALUES (10189,'27b70837','a4c7b9dd136f5f44de51da62f08b94dab937d34060d92ea5aa332b7165d39840','27b70837f21ed26d92937f6d2827feb0a0433158@email.com','name','last','2022-06-01 13:22:51','-',0,2,1,'HOIBLUCOXDJDUUVB',NULL,NULL,NULL,0,'2022-06-01 13:22:51',NULL,10,2,0,0,0,'EvBV3f0ttNNJ6/ODHdRD9OaVJL6yQcGBQ7YZ+afDq9hHxcLV9MgIkq2LqtIIRaoEy/WJfySsM3jvfETObDr/YQ==');",
			"INSERT INTO `users` VALUES (10191,'5a93543a','8a7f9c810a3fb66ccf929b7b71fc20484e8a31a060b04608cda7b334402133dd','5a93543a6f6ef2964aa84ec23fbd750b4868c2b4@email.com','name','last','2022-06-01 13:30:43','-',0,2,1,'WIYJ5CPQFVA7FDLT',NULL,NULL,NULL,0,'2022-06-01 13:30:43',NULL,10,2,0,0,0,'MehzrCMfoQSQUJccmLOGnytKE5S9ocM8/ZzuXpfBfgyfB+yKSN3Zxnq29fsHSEbPp1b6oizYy8JjhE9xceqaAw==');",
			"INSERT INTO `users` VALUES (10001,'agent007','fcb581887fcf61bf1a9e3ddd2f64297a9179efdd7ea32443021ea72e1f232b26','agent@hops.io','Agent','007','2015-05-15 10:22:36','Mrs',-1,2,0,'V3WBPS4G2WMQ53VA',NULL,NULL,NULL,0,'2015-04-28 15:18:42',NULL,0,0,0,1,3,'+9mTLmYSpnZROFEJEaednw8+gDH/s2J1QuRZy8okxW5myI/q8ek8Xu+ab5CyE9GzhWX6Sa4cr7KX8cAHi5IC4g==');",
			"INSERT INTO `users` VALUES (10181,'2e32360d','c30f9834b1b935e7217e2fa647b45837153820ef06a2796cf1218687b800c9cd','2e32360dd892452068a0fc1128b0bb36bd570c84@email.com','name','last','2022-06-01 13:19:50','-',0,2,0,'TYAPQOQEZLM7VBG3',NULL,NULL,NULL,0,'2022-06-01 13:19:49',NULL,10,0,0,0,0,'aaVNEU0HejWSGYH3Hmtz8f5fRZKYygwB5iKioSFnAwu8EDrGRstedB5XsB2QgshEklxkpqtZrZ8N84ZE23iiug==');",
			"INSERT INTO `users` VALUES (10184,'46005dd1','571b044e29ed8f55f16cea778d5494366d1ede825d32bb7b10d42992e9d6a310','46005dd1690ad878c8b01109be3bcc7eaf1509ff@email.com','name','last','2022-06-01 13:20:25','-',0,2,0,'W367MZLNQYQUMGL3',NULL,NULL,NULL,0,'2022-06-01 13:20:24',NULL,10,0,0,0,0,'21rdaaVWDVCIVN15OemD6P9VTQxdHV5DPQbrNC57UnEl1qLcXniiVrEkPwEGuNnU0Z3IEi3K+FsPjlBDpQpg2A==');",
			"INSERT INTO `users` VALUES (10186,'user17f0','3dee2df7285d35595ab27e92ddd112246d3052186f23b5f566cdcaacd4ac21fb','user1_7f0b9b5ca61be1595dc7464097a6f3399344f2ed@email.com','User','1','2022-06-01 13:20:28','-',0,2,1,'L5OGREUPR26RXZTO',NULL,NULL,NULL,0,'2022-06-01 13:20:28',NULL,10,2,0,0,0,'CyGVqie5ICNXwi7kIHWqSZuRJjBlcSkMx/YCKn83hJCCWnJ2pGBfJMTKSSzLBNep9RLcKwKjh5UCG5uWyp0pjQ==');",
			"INSERT INTO `users` VALUES (10187,'user21b6','3e1b486eac35582bf6cfcc4f8bfd54652e0465d25ec0a5beee6afcf2db4a1089','user2_1b67f8d48fd22a5f6fa687f7f2eb3d4019fbbd16@email.com','User','2','2022-06-01 13:20:28','-',0,2,1,'KGCP2KACX3LU7VJT',NULL,NULL,NULL,0,'2022-06-01 13:20:28',NULL,10,2,0,0,0,'bvvm1u5sVNAWGiknb0DYzAALsL/C8G+nS1I9fE87mzf70mxsLAXyfLpag9RQFu017sLMu25bMyjV0laablL8tg==');",
			"INSERT INTO `users` VALUES (10190,'7304b930','3b180177f574cdd841f9e3d898677efbb32f75f3f32e002fd49bed60acacb459','7304b93005c4ab8fb16de01e0fafe4bc968b377e@email.com','name','last','2022-06-01 13:26:09','-',0,2,1,'6FZ4C56HBOAYH6EM',NULL,NULL,NULL,0,'2022-06-01 13:26:09',NULL,10,0,1,0,0,'60SQFTuavPRcxPphQPCdE/Ja757dzwc49nCL4B+kjiVvzh9WxpLVYHKJEJn9aAQFU4kNSnqy34g+CFmpCHys4g==')",

			"INSERT INTO `project` VALUES (119,322,'demo_fs_meb10000',322,'demo_fs_meb10000','admin@hopsworks.ai'," +
				"'2022-05-30 14:17:22','2032-05-30',0,0,NULL,'A demo project for getting started with featurestore'," +
				"'NOLIMIT','2022-05-30 14:17:38',100,'demo_fs_meb10000:1653921933268-2.6.0-SNAPSHOT.1',1)",
			"INSERT INTO `project` VALUES (128,322,'online_fs1',322,'online_fs1'," +
				"'7304b93005c4ab8fb16de01e0fafe4bc968b377e@email.com','2022-06-01 13:27:50','2032-06-01',0,0,NULL,''" +
				",'NOLIMIT','2022-06-01 13:28:10',100,'python38:2.6.0-SNAPSHOT',10)",

			"INSERT INTO `project_team` VALUES (128,'7304b93005c4ab8fb16de01e0fafe4bc968b377e@email.com','Data owner','2022-06-01 13:27:51')",
			"INSERT INTO `project_team` VALUES (128,'onlinefs@hopsworks.ai','Data scientist','2022-06-01 13:28:17')",
			"INSERT INTO `project_team` VALUES (119,'admin@hopsworks.ai','Data owner','2022-05-30 14:17:24')",
			"INSERT INTO `project_team` VALUES (119,'onlinefs@hopsworks.ai','Data scientist','2022-05-30 14:17:43')",
			"INSERT INTO `project_team` VALUES (128,'serving@hopsworks.se','Data scientist','2022-06-01 13:28:05')",

			// 1  bkYjEz6OTZyevbqt.ocHajJhnE0ytBh8zbYj3IXupyMqeMZp8PW464eTxzxqP5afBjodEQUgY0lmL33ub
			// 2  oqZdmmRYy5QCwh55.nbqiMNJS3BHMe1OabnaZtAH8OiU39A2DFNP0WSU8LNhjwJgBnaAkU4veXLTi4bmy
			// 3  TsbTJMYyobErsbXY.bgi8GIsey3hkTzswyVSRm1B05qXoDuz55M6WXiSHwfiqxg7i9RgJ20Wz5ZFW9h7b

			"INSERT INTO `api_key` VALUES (2049 , 'bkYjEz6OTZyevbqt' , '709faa77accc3f30394cfb53b67253ba64881528cb3056eea110703ca430cce4' , '1/1TxiaiIB01rIcY2E36iuwKP6fm2GzBaNaQqOVGMhH0AvcIlIzaUIw0fMDjKNLa0OWxAOrfTSPqAolpI/n+ug==' , '2022-06-14 10:27:03' , '2022-06-14 10:27:03' , 'myapikey1'             ,   10179 ,        0 )",
			"INSERT INTO `api_key` VALUES (2050 , 'oqZdmmRYy5QCwh55' , '649df9aed4609d4c65a5e0d519eed754e3dc97a65df22133a547a1cee7a8e4a6' , 'o24ewhFzfXzulNwII9oGPbRwPZP/UYcBfuyD3HQEECATHERdcCz4h8owbB3rMhbQtXNPRHw1uY1F+l7GAiKKow==' , '2022-06-14 10:27:44' , '2022-06-14 10:27:44' , 'myapikey2'             ,   10000 ,        0 )",
			"INSERT INTO `api_key` VALUES (2051 , 'TsbTJMYyobErsbXY' , 'f790ffc635b4251283317caf888b0e44a76c8fedd234e5a975938654c15b91c9' , 'k0hcvIpu66ojKEU7CLY+3xNqVD3Zj7cHWyb/ioaasMrbEyo1VBPZ2oqg01mOoEmMfmZ7v9cjOMez/bC0dK/sZw==' , '2022-06-14 10:28:13' , '2022-06-14 10:28:13' , 'myapikey3'             ,   10179 ,        0 )",
		},

		{ // clean up commands
			// "DROP DATABASE " + db,
		},
	}

}

func SchemaTextualColumns(colType string, db string, length int) [][]string {
	if strings.EqualFold(colType, "varbinary") || strings.EqualFold(colType, "binary") ||
		strings.EqualFold(colType, "char") || strings.EqualFold(colType, "varchar") {
		return [][]string{
			{
				// setup commands
				"DROP DATABASE IF EXISTS " + db,
				"CREATE DATABASE " + db,
				"USE " + db,

				// blobs in PK is not supported by RonDB
				"CREATE TABLE table1(id0 " + colType + "(" + strconv.Itoa(length) + "), col0 " + colType + "(" + strconv.Itoa(length) + "),  PRIMARY KEY(id0))",
				`INSERT INTO  table1 VALUES("1","这是一个测验。 我不知道怎么读中文。")`,
				`INSERT INTO  table1 VALUES("2",0x660066)`,
				`INSERT INTO  table1 VALUES("3","a\nb")`,
				`INSERT INTO  table1 VALUES("这是一个测验","12345")`,
				`INSERT INTO  table1 VALUES("4","ÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖØÙÚÛÜÝÞßàáâãäåæçèéêëìíîïð")`, // some chars
				`INSERT INTO  table1 set id0=5`,
				`INSERT INTO  table1 VALUES("6","\"\\\b\f\n\r\t$%_?")`, // in mysql \f is replaced by f
			},

			{ // clean up commands
				"DROP DATABASE " + db,
			},
		}
	} else {
		panic("Data type not supported")
	}
}

func benchmarSchema(db string, count int) []string {
	colWidth := 1000
	dummyData := ""
	for i := 0; i < colWidth; i++ {
		dummyData += "$"
	}

	schema := []string{
		// setup commands
		"DROP DATABASE IF EXISTS " + db,
		"CREATE DATABASE " + db,
		"USE " + db,
		"CREATE TABLE table_1(id0 INT, col_0 VARCHAR(" + strconv.Itoa(colWidth) + "), PRIMARY KEY(id0))",
	}

	for i := 0; i < count; i++ {
		schema = append(schema, fmt.Sprintf("INSERT INTO table_1 VALUES(%d, \"%s\")", i, dummyData))
	}

	return schema
}

func Database(name string) [][]string {
	db, ok := databases[name]
	if !ok {
		return [][]string{}
	}
	return db
}
