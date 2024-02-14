from database import Database


def seed_users():
    db = Database()
    conn = db.get_connection()
    conn.execute("INSERT OR IGNORE INTO users (username, password) VALUES ('admin', 'password')")
    conn.execute("INSERT OR IGNORE INTO users (username, password) VALUES ('user', 'password')")
    conn.commit()
    conn.close()


def main():
    seed_users()


if __name__ == "__main__":
    main()
