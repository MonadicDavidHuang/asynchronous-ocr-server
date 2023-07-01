-- CreateTable
CREATE TABLE `tasks` (
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `open_task_id` VARCHAR(26) NOT NULL,
    `task_status` ENUM('pending', 'complete', 'deleted') NOT NULL DEFAULT 'pending',
    `image_file_id` int(11),
    `image_file_status` ENUM('uploaded', 'deleted') NOT NULL DEFAULT 'uploaded',
    `caption` MEDIUMTEXT,
    `created_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    `updated_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (`id`),
    UNIQUE KEY `open_task_id_index`(`open_task_id`),
    KEY `task_status_index`(`task_status`),
    KEY `image_file_status_index`(`image_file_status`)
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;