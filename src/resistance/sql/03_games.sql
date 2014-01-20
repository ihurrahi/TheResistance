# Descibes the games table that stores each game

CREATE TABLE IF NOT EXISTS `games` (
  `game_id` BIGINT(20) NOT NULL AUTO_INCREMENT,
  `title` VARCHAR(30) NOT NULL,
  `host_id` BIGINT(20) NOT NULL,
  `status` CHAR(1) NOT NULL,
  PRIMARY KEY(`game_id`)
)
