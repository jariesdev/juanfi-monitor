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


def create_juanfi_logs_table():
    conn = sqlite3.connect("app.db")
    conn.execute(
        "CREATE TABLE IF NOT EXISTS juanfi_logs("
        "id integer primary key AUTOINCREMENT not null, "
        "log_time char(50), "
        "description text, "
        "created_at char(50) DEFAULT CURRENT_TIMESTAMP, "
        "UNIQUE (log_time))")
    conn.commit()
    conn.close()
    print("juanfi_logs table created successfully")

def create_juanfi_sales_table():
    conn = sqlite3.connect("app.db")
    conn.execute(
        "CREATE TABLE IF NOT EXISTS juanfi_sales("
        "id integer primary key AUTOINCREMENT not null, "
        "sale_time char(50), "
        "mac_address char(50), "
        "voucher char(20), "
        "amount real, "
        "created_at char(50) DEFAULT CURRENT_TIMESTAMP, "
        "UNIQUE (sale_time))")
    conn.commit()
    conn.close()
    print("juanfi_sales table created successfully")


def main():
    print("Start migration")
    create_user_table()
    create_juanfi_logs_table()
    create_juanfi_sales_table()


if __name__ == "__main__":
    main()
