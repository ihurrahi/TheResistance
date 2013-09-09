# Describes the teams table - indicating which players went on which mission

CREATE TABLE IF NOT EXISTS `teams` (
  `team_id` BIGINT(20) NOT NULL AUTO_INCREMENT,
  `mission_id` BIGINT(20) NOT NULL,
  `user_id` BIGINT(20) NOT NULL,
  PRIMARY KEY (`team_id`)
)