import pymssql

conn = pymssql.connect(server='mssql', user='sa', password='YourStrong@Passw0rd123', database='api')
cursor = conn.cursor()

with open('/backup/db_export_structure.sql', 'w', encoding='utf-8') as f:
    
    # 1. EXPORT CREATE TABLES (DDL)
    f.write("-- ==========================================\n")
    f.write("-- CREATE TABLES (SCHEMA)\n")
    f.write("-- ==========================================\n\n")
    
    cursor.execute("SELECT TABLE_SCHEMA, TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE = 'BASE TABLE'")
    tables = cursor.fetchall()
    
    for schema_name, table_name in tables:
        f.write(f"-- Tabel: [{schema_name}].[{table_name}]\n")
        f.write(f"CREATE TABLE [{schema_name}].[{table_name}] (\n")
        
        # Ambil definisi kolom
        cursor.execute("""
            SELECT 
                COLUMN_NAME,
                DATA_TYPE,
                CHARACTER_MAXIMUM_LENGTH,
                NUMERIC_PRECISION,
                NUMERIC_SCALE,
                IS_NULLABLE,
                COLUMN_DEFAULT
            FROM INFORMATION_SCHEMA.COLUMNS
            WHERE TABLE_SCHEMA = %s AND TABLE_NAME = %s
            ORDER BY ORDINAL_POSITION
        """, (schema_name, table_name))
        columns = cursor.fetchall()
        col_defs = []
        
        for col in columns:
            c_name, c_type, char_len, num_prec, num_scale, is_null, c_def = col
            type_str = c_type.upper()
            
            # Penanganan panjang tipe data string/binary
            if type_str in ('VARCHAR', 'NVARCHAR', 'CHAR', 'NCHAR', 'VARBINARY', 'BINARY'):
                if char_len == -1:
                    type_str += "(MAX)"
                elif char_len:
                    type_str += f"({char_len})"
            elif type_str in ('DECIMAL', 'NUMERIC'):
                if num_prec is not None and num_scale is not None:
                    type_str += f"({num_prec},{num_scale})"
                    
            null_str = "NULL" if is_null == "YES" else "NOT NULL"
            def_str = f" DEFAULT {c_def}" if c_def else ""
            
            col_defs.append(f"    [{c_name}] {type_str} {null_str}{def_str}")
            
        f.write(",\n".join(col_defs))
        f.write("\n);\nGO\n\n")

    
    # 3. EXPORT STORED PROCEDURES
    f.write("-- ==========================================\n")
    f.write("-- STORED PROCEDURES\n")
    f.write("-- ==========================================\n\n")
    
    cursor.execute("""
        SELECT 
            s.name AS schema_name,
            p.name AS proc_name,
            m.definition AS definition
        FROM sys.procedures p
        INNER JOIN sys.sql_modules m ON p.object_id = m.object_id
        INNER JOIN sys.schemas s ON p.schema_id = s.schema_id
    """)
    procs = cursor.fetchall()
    
    for schema_name, proc_name, definition in procs:
        f.write(f"-- Stored Procedure: [{schema_name}].[{proc_name}]\n")
        f.write("GO\n")
        if definition:
            f.write(definition.strip() + "\n")
        f.write("GO\n\n")

conn.close()
print("Export lengkap (CREATE TABLE + INSERT Data + Stored Procedures) berhasil!")