# Adds the role column to the players table. This column
# is for whether the player is on the resistance or spy team.

ALTER TABLE `players` ADD `role` TINYINT(3) DEFAULT NULL;
