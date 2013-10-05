# Describes the players table - indicating which players played in which game

CREATE TABLE IF NOT EXISTS `players` (
  `game_id` BIGINT(20) NOT NULL,
  `user_id` BIGINT(20) NOT NULL,
  PRIMARY KEY (game_id,user_id)
)