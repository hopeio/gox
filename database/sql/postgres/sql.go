/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package postgres

const (
	// Create a function that drops tables in the public schema for the given username
	DeleteTablesFunc = `CREATE OR REPLACE FUNCTION del_tabs(username IN VARCHAR) RETURNS void AS $$
DECLARE
	statements CURSOR FOR
		SELECT tablename FROM pg_tables
		WHERE tableowner = username AND schemaname = 'public';
BEGIN
	FOR stmt IN statements LOOP
		EXECUTE 'DROP TABLE ' || quote_ident(stmt.tablename) || ' CASCADE;';
	END LOOP;
END;
$$ LANGUAGE plpgsql`
	DeleteTables = `SELECT del_tabs('postgres')`
	// Create a function that truncates tables in the public schema for the given username
	TruncateTablesFunc = `CREATE OR REPLACE FUNCTION truncate_tables(username IN VARCHAR) RETURNS void AS $$
DECLARE
    statements CURSOR FOR
        SELECT tablename FROM pg_tables
        WHERE tableowner = username AND schemaname = 'public';
BEGIN
    FOR stmt IN statements LOOP
        EXECUTE 'TRUNCATE TABLE ' || quote_ident(stmt.tablename) || ' CASCADE;';
    END LOOP;
END;
$$ LANGUAGE plpgsql;`
	TruncateTables = `SELECT truncate_tables('postgres')`
)
