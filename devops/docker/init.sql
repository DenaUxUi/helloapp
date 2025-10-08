CREATE TABLE IF NOT EXISTS instances (
      id SERIAL PRIMARY KEY,
      name TEXT,
      created_at TIMESTAMP DEFAULT now()
    );  