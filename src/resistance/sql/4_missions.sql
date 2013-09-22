# Describes the missions table that stores each mission.
# Up to 5 per game

CREATE TABLE IF NOT EXISTS `missions` (
  `mission_id` BIGINT(20) NOT NULL AUTO_INCREMENT,
  `game_id` BIGINT(20) NOT NULL,
  `mission_num` INT(5) NOT NULL,
  `leader_id` BIGINT(20) NOT NULL,
  `winner` VARCHAR(30) NOT NULL,
  `num_fails` INT(5) NOT NULL,
  PRIMARY KEY (`mission_id`)
)