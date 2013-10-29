# Creates the votes table that stores the votes for each player for a specific team

CREATE TABLE IF NOT EXISTS `votes` (
  `team_id` BIGINT(20) NOT NULL,
  `user_id` BIGINT(20) NOT NULL,
  `vote` BIT(1) NOT NULL,
  PRIMARY KEY(`team_id`, `user_id`)
)