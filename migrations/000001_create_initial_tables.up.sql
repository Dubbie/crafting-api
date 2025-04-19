CREATE TABLE items (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    slug VARCHAR(255) NOT NULL UNIQUE,
    is_raw_material BOOLEAN NOT NULL DEFAULT false,
    description TEXT NULL,
    image_url VARCHAR(255) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_items_slug (slug),
    INDEX idx_items_is_raw_material (is_raw_material)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add CREATE TABLE for crafting_methods here...
CREATE TABLE crafting_methods (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_crafting_methods_slug (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add CREATE TABLE for recipes here (including FK constraint)...
CREATE TABLE recipes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NULL UNIQUE, -- Nullable name initially
    crafting_method_id INT UNSIGNED NOT NULL,
    eu_per_tick INT UNSIGNED NULL,
    duration_ticks INT UNSIGNED NULL,
    notes TEXT NULL,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_recipes_is_default (is_default),
    CONSTRAINT fk_recipes_crafting_method
        FOREIGN KEY (crafting_method_id) REFERENCES crafting_methods(id)
        ON DELETE RESTRICT -- Or CASCADE, depending on desired behavior
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add CREATE TABLE for recipe_inputs...
CREATE TABLE recipe_inputs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    recipe_id BIGINT UNSIGNED NOT NULL,
    input_item_id BIGINT UNSIGNED NOT NULL,
    input_quantity INT UNSIGNED NOT NULL,
    UNIQUE KEY uq_recipe_input (recipe_id, input_item_id),
    CONSTRAINT fk_recipe_inputs_recipe
        FOREIGN KEY (recipe_id) REFERENCES recipes(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_recipe_inputs_item
        FOREIGN KEY (input_item_id) REFERENCES items(id)
        ON DELETE CASCADE -- Cascade delete if an item is deleted? Or RESTRICT?
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add CREATE TABLE for recipe_outputs...
CREATE TABLE recipe_outputs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    recipe_id BIGINT UNSIGNED NOT NULL,
    item_id BIGINT UNSIGNED NOT NULL,
    quantity INT UNSIGNED NOT NULL DEFAULT 1,
    chance INT UNSIGNED NOT NULL DEFAULT 10000, -- Represents 100.00%
    is_primary_output BOOLEAN NOT NULL DEFAULT false,
    UNIQUE KEY uq_recipe_output (recipe_id, item_id),
    INDEX idx_recipe_outputs_is_primary (is_primary_output),
    CONSTRAINT fk_recipe_outputs_recipe
        FOREIGN KEY (recipe_id) REFERENCES recipes(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_recipe_outputs_item
        FOREIGN KEY (item_id) REFERENCES items(id)
        ON DELETE CASCADE -- Cascade delete if an item is deleted? Or RESTRICT?
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add CREATE TABLE for users... (Example, adjust as needed)
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL, -- Store hashed passwords!
    remember_token VARCHAR(100) NULL,
    created_at TIMESTAMP NULL,
    updated_at TIMESTAMP NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add CREATE TABLE for user_preferred_recipes...
CREATE TABLE user_preferred_recipes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    output_item_id BIGINT UNSIGNED NOT NULL,
    preferred_recipe_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_user_output_item (user_id, output_item_id),
    CONSTRAINT fk_user_preferred_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_user_preferred_output_item
        FOREIGN KEY (output_item_id) REFERENCES items(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_user_preferred_recipe
        FOREIGN KEY (preferred_recipe_id) REFERENCES recipes(id)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
