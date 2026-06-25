CREATE TABLE IF NOT EXISTS books(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    book TEXT NOT NULL,
    rating INT NOT NULL,
    start_date DATE,
    finish_date DATE,
    pages INT,
    thoughts TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
