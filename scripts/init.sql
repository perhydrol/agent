-- Users Table
CREATE TABLE IF NOT EXISTS `users` (
    `id` BIGINT NOT NULL,
    `username` VARCHAR(32) NOT NULL,
    `password_hash` VARCHAR(128) NOT NULL,
    `email` VARCHAR(128) NOT NULL,
    `created_at` DATETIME(3) NULL,
    `updated_at` DATETIME(3) NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_users_username` (`username`),
    UNIQUE INDEX `idx_users_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Products Table
CREATE TABLE IF NOT EXISTS `products` (
    `id` BIGINT NOT NULL,
    `name` VARCHAR(128) NOT NULL,
    `category` VARCHAR(32) NOT NULL,
    `base_price` DECIMAL(10,2) NOT NULL,
    `description` TEXT,
    `features` JSON NOT NULL,
    `created_at` DATETIME(3) NULL,
    `updated_at` DATETIME(3) NULL,
    PRIMARY KEY (`id`),
    INDEX `idx_products_category` (`category`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Orders Table
CREATE TABLE IF NOT EXISTS `orders` (
    `id` BIGINT NOT NULL,
    `user_id` BIGINT NOT NULL,
    `product_id` BIGINT NOT NULL,
    `product_name_snapshot` VARCHAR(128) NOT NULL,
    `product_price_snapshot` DECIMAL(10,2) NOT NULL,
    `total_amount` DECIMAL(10,2) NOT NULL,
    `status` TINYINT DEFAULT 0,
    `policy_number` VARCHAR(64),
    `version` BIGINT DEFAULT 1,
    `created_at` DATETIME(3) NULL,
    `updated_at` DATETIME(3) NULL,
    PRIMARY KEY (`id`),
    INDEX `idx_orders_user_id` (`user_id`),
    INDEX `idx_orders_product_id` (`product_id`),
    INDEX `idx_orders_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Chat Messages Table
CREATE TABLE IF NOT EXISTS `chat_messages` (
    `id` BIGINT AUTO_INCREMENT,
    `session_id` VARCHAR(64) NOT NULL,
    `user_id` BIGINT NOT NULL,
    `role` VARCHAR(10) NOT NULL,
    `content` TEXT NOT NULL,
    `created_at` DATETIME(3) NULL,
    PRIMARY KEY (`id`),
    INDEX `idx_chat_messages_session_id` (`session_id`),
    INDEX `idx_chat_messages_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Initial Data for Products
INSERT IGNORE INTO `products` (`id`, `name`, `category`, `base_price`, `description`, `features`, `created_at`, `updated_at`) VALUES
(1001, 'Travel Safe Basic', 'Travel', 50.00, 'Basic travel insurance covering medical expenses.', '{"coverage": "50000 USD", "duration": "30 days"}', NOW(), NOW()),
(1002, 'Travel Safe Premium', 'Travel', 120.00, 'Premium travel insurance with flight cancellation.', '{"coverage": "100000 USD", "duration": "30 days", "cancellation": true}', NOW(), NOW()),
(1003, 'Health Plus', 'Health', 200.00, 'Comprehensive health insurance.', '{"coverage": "Unlimited", "deductible": "500 USD"}', NOW(), NOW());
