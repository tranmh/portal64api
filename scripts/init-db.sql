-- Database initialization script for Portal64 API
-- This script creates the necessary databases and users for development

-- Create databases
CREATE DATABASE IF NOT EXISTS mvdsb CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS portal64_bdw CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS portal64_svw CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create user and grant permissions
CREATE USER IF NOT EXISTS 'portal64api'@'%' IDENTIFIED BY 'password123';

-- Grant permissions to all databases
GRANT SELECT, INSERT, UPDATE, DELETE ON mvdsb.* TO 'portal64api'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON portal64_bdw.* TO 'portal64api'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON portal64_svw.* TO 'portal64api'@'%';

-- For testing purposes, also grant to localhost
GRANT SELECT, INSERT, UPDATE, DELETE ON mvdsb.* TO 'portal64api'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON portal64_bdw.* TO 'portal64api'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON portal64_svw.* TO 'portal64api'@'localhost';

-- Flush privileges
FLUSH PRIVILEGES;

-- Create some sample data for testing (optional)
USE mvdsb;

-- Sample organizations (clubs)
INSERT IGNORE INTO organisation (id, name, kurzname, vkz, verband, unterverband, bezirk, verein, organisationsart, status) VALUES
(1, 'Post-SV Ulm', 'Post-SV Ulm', 'C0101', 'W', '01', '01', '01', 20, 1),
(2, 'Schachfreunde Stuttgart', 'SF Stuttgart', 'C0201', 'W', '02', '01', '01', 20, 1),
(3, 'TSV Bayer Dormagen', 'TSV Dormagen', 'C0301', 'B', '03', '01', '01', 20, 1);

-- Sample persons (players)
INSERT IGNORE INTO person (id, name, vorname, geburtsdatum, geschlecht, nation, status) VALUES
(1014, 'Sick', 'Oliver', '1980-05-15', 1, 'GER', 1),
(1015, 'Müller', 'Hans', '1975-03-22', 1, 'GER', 1),
(1016, 'Schmidt', 'Anna', '1982-11-08', 2, 'GER', 1),
(1017, 'Weber', 'Michael', '1978-07-12', 1, 'GER', 1),
(1018, 'Fischer', 'Emma', '1990-01-25', 2, 'GER', 1);

-- Sample memberships
INSERT IGNORE INTO mitgliedschaft (id, person, organisation, von, stat1, stat2, status) VALUES
(1, 1014, 1, '2020-01-01', 1, 1, 1),
(2, 1015, 1, '2019-06-15', 1, 1, 1),
(3, 1016, 2, '2021-03-01', 1, 1, 1),
(4, 1017, 2, '2018-09-10', 1, 1, 1),
(5, 1018, 3, '2022-01-15', 1, 1, 1);

-- Switch to Portal64_BDW database for evaluations
USE portal64_bdw;

-- Sample tournaments
INSERT IGNORE INTO tournament (id, tname, tcode, acron, type, rounds, finishedOn, idOrganisation) VALUES
(1, 'Ulmer Stadtmeisterschaft 2024', 'ULM24-SM', 'USM', 'RR', 9, '2024-12-15', 1),
(2, 'Stuttgart Open 2024', 'STU24-OP', 'SOP', 'SW', 7, '2024-11-20', 2);

-- Sample evaluations (DWZ ratings)
INSERT IGNORE INTO evaluation (id, idMaster, idPerson, dwzOld, dwzOldIndex, dwzNew, dwzNewIndex, games, points) VALUES
(1, 1, 1014, 2140, 42, 2156, 45, 7, 4.5),
(2, 1, 1015, 1856, 38, 1862, 41, 7, 3.0),
(3, 2, 1016, 1654, 22, 1671, 25, 6, 4.0),
(4, 2, 1017, 1923, 55, 1918, 58, 6, 2.5),
(5, 1, 1018, 1445, 15, 1465, 18, 7, 3.5);

-- Switch to Portal64_SVW database for tournaments
USE portal64_svw;

-- Sample SVW tournaments
INSERT IGNORE INTO Turnier (TID, TName, Saison, SaisonAnzeige, AnzStammspieler, AnzErsatzspieler, Teilnahmeschluss, Meldeschluss, isFreigegeben, Organisation) VALUES
(1, 'Württembergische Verbandsliga 2025', 2025, '2024/25', 8, 2, '2025-02-15 23:59:59', '2025-01-31 23:59:59', 1, 1),
(2, 'Bezirksliga Stuttgart 2025', 2025, '2024/25', 8, 2, '2025-02-20 23:59:59', '2025-02-05 23:59:59', 1, 2);

-- Add some sample teams and players for tournaments
INSERT IGNORE INTO Mannschaft (MFTID, VKZ, TID, MName, MNr, Los, organisation) VALUES
(1, 'C0101', 1, 'Post-SV Ulm I', 1, 1, 1),
(2, 'C0201', 1, 'SF Stuttgart I', 1, 2, 2);

-- Add sample team members
INSERT IGNORE INTO Kader (KID, MFTID, person, Brettnr, abgemeldet) VALUES
(1, 1, 1014, 1, 0),
(2, 1, 1015, 2, 0),
(3, 2, 1016, 1, 0),
(4, 2, 1017, 2, 0);

COMMIT;
