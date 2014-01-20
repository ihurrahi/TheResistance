# Adds the outcome column to the teams for storing a bit representing a
# success or fail for the mission the team is associated with.

ALTER TABLE `teams` ADD `outcome` CHAR(1) DEFAULT NULL;
