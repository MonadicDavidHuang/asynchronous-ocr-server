-- CreateTable
CREATE TABLE `image_files` (
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `content` BLOB NOT NULL,
    `file_type` ENUM('image/jpeg', 'image/png', 'image/tiff') NOT NULL DEFAULT 'image/jpeg',
    `created_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    `updated_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (`id`)
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;