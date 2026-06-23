-- Membuat tabel sample untuk tugas
CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    project_name VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Memasukkan data awal (seeding)
INSERT INTO projects (project_name, status) VALUES
('Dockerize Node.js API', 'Completed'),
('Dockerize Python Flask', 'Completed'),
('Dockerize React SPA', 'Completed'),
('Dockerize Go Binary', 'Completed'),
('Dockerize PostgreSQL Custom', 'In Progress');
