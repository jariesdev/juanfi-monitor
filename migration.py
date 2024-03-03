import sqlite3


def create_user_table():
    conn = sqlite3.connect("app.db")
    conn.execute("CREATE TABLE IF NOT EXISTS users("
                 "id integer primary key AUTOINCREMENT not null, "
                 "username char(50), "
                 "password text, "
                 "UNIQUE (username))")
    conn.commit()
    conn.close()
    print("users table created successfully")


def create_vendo_logs_table():
    conn = sqlite3.connect("app.db")
    conn.execute(
        "CREATE TABLE IF NOT EXISTS vendo_logs("
        "id integer primary key AUTOINCREMENT not null, "
        "vendo_id integer, "
        "log_time char(50), "
        "description text, "
        "created_at char(50) DEFAULT CURRENT_TIMESTAMP, "
        "UNIQUE (log_time))")
    conn.commit()
    conn.close()
    print("vendo_logs table created successfully")

def create_vendo_sales_table():
    conn = sqlite3.connect("app.db")
    conn.execute(
        "CREATE TABLE IF NOT EXISTS vendo_sales("
        "id integer primary key AUTOINCREMENT not null, "
        "vendo_id integer, "
        "sale_time char(50), "
        "mac_address char(50), "
        "voucher char(20), "
        "amount real, "
        "created_at char(50) DEFAULT CURRENT_TIMESTAMP, "
        "UNIQUE (sale_time))")
    conn.commit()
    conn.close()
    print("vendo_sales table created successfully")

def create_vendo_status_table():
    conn = sqlite3.connect("app.db")
    conn.execute(
        "CREATE TABLE IF NOT EXISTS vendo_status("
        "id integer primary key AUTOINCREMENT not null, "
        "vendo_id integer, "
        "is_online integer, "
        "active_users integer, "
        "wireless_strength integer, "
        "created_at char(50) DEFAULT CURRENT_TIMESTAMP)")
    conn.commit()
    conn.close()
    print("vendos table created successfully")

def create_vendos_table():
    conn = sqlite3.connect("app.db")
    conn.execute(
        "CREATE TABLE IF NOT EXISTS vendos("
        "id integer primary key AUTOINCREMENT not null, "
        "name char(50), "
        "mac_address char(50), "
        "api_url char(50), "
        "api_key char(20), "
        "is_online integer, "
        "total_sales integer, "
        "current_sales integer, "
        "created_at char(50) DEFAULT CURRENT_TIMESTAMP)")
    conn.commit()
    conn.close()
    print("vendos table created successfully")


def main():
    print("Start migration")
    create_user_table()
    create_vendo_logs_table()
    create_vendo_sales_table()
    create_vendo_status_table()
    create_vendos_table()


if __name__ == "__main__":
    main()
