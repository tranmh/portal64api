-- Portal64API Performance Fix
-- Add composite indexes that include status field for efficient filtering

USE mvdsb;

-- Add composite index for status + name searches
-- This will allow MySQL to filter by status=0 first, then search names
CREATE INDEX idx_person_status_name ON person(status, name);

-- Add composite index for status + vorname searches  
-- This handles firstname searches efficiently
CREATE INDEX idx_person_status_vorname ON person(status, vorname);

-- Add composite index for mitgliedschaft lookups with status
-- This improves club membership queries: person + bis IS NULL + status = 0
CREATE INDEX idx_mitgliedschaft_person_bis_status ON mitgliedschaft(person, bis, status);

-- Verify the new indexes were created
SHOW INDEX FROM person WHERE Key_name LIKE 'idx_%';
SHOW INDEX FROM mitgliedschaft WHERE Key_name LIKE 'idx_%';

-- Test the query performance after adding indexes
EXPLAIN SELECT count(*) FROM person WHERE status = 0 AND ((name >= 'Tran' AND name < 'Tranzz') OR (vorname >= 'Tran' AND vorname < 'Tranzz'));
