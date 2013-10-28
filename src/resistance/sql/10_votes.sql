# Creates the votes table that stores the votes for each player for a specific team

CREATE TABLE IF NOT EXISTS `votes` (
  `vote_id` BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `team_id` BIGINT(20) UNSIGNED NOT NULL,
  `user_id` BIGINT(20) UNSIGNED NOT NULL,
  `vote` BIT(1) DEFAULT NULL,
  PRIMARY KEY(`vote_id`)
)