# Describes the teams table - indicating which players went on which mission

CREATE TABLE IF NOT EXISTS `teams` (
  `mission_id` BIGINT(20) NOT NULL,
  `user_id` BIGINT(20) NOT NULL,
  PRIMARY KEY (`mission_id`, `user_id`)
)