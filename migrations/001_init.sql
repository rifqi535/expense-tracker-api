-- users
CREATE TABLE IF NOT EXISTS users (
id UUID PRIMARY KEY,
name TEXT NOT NULL,
email TEXT NOT NULL UNIQUE,
password_hash TEXT NOT NULL,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- categories
CREATE TABLE IF NOT EXISTS categories (
id UUID PRIMARY KEY,
title TEXT NOT NULL,
user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
CONSTRAINT uq_user_category UNIQUE (user_id, title)
);
CREATE INDEX IF NOT EXISTS idx_categories_user ON categories(user_id);


-- expenses
CREATE TABLE IF NOT EXISTS expenses (
id UUID PRIMARY KEY,
title TEXT NOT NULL,
description TEXT,
amount NUMERIC(12,2) NOT NULL CHECK (amount > 0),
category_id UUID NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()

);
CREATE INDEX IF NOT EXISTS idx_expenses_user ON expenses(user_id);
CREATE INDEX IF NOT EXISTS idx_expenses_category ON expenses(category_id);
CREATE INDEX IF NOT EXISTS idx_expenses_created_at ON expenses(created_at);
