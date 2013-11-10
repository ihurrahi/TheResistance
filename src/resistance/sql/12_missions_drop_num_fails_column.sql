# Drops the num_fails column in the missions table. The source of truth
# on the number of fails should be on the teams table that points to the mission.

ALTER TABLE `missions` DROP COLUMN `num_fails`;