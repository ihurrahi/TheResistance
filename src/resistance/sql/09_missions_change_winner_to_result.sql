# Changes the winner column to result to account for canceled missions

ALTER TABLE `missions` CHANGE `winner` `result` VARCHAR(30) NOT NULL;