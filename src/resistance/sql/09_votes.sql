# Creates the votes table that stores the votes for each player for a specific team

CREATE TABLE IF NOT EXISTS `votes` (
  `mission_id` BIGINT(20) NOT NULL,
  `user_id` BIGINT(20) NOT NULL,
  `vote` CHAR(1) DEFAULT NULL,
  PRIMARY KEY(`mission_id`, `user_id`)
)
