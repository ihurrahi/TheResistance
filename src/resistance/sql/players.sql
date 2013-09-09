# Describes the players table - indicating which players played in which game

CREATE TABLE IF NOT EXISTS `players` (
  `players_id` BIGINT(20) NOT NULL AUTO_INCREMENT,
  `game_id` BIGINT(20) NOT NULL,
  `user_id` BIGINT(20) NOT NULL,
  PRIMARY KEY (players_id)
)