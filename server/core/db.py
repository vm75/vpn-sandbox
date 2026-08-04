import sqlite3
import json
import os

Db = None

def get_db():
    return Db


def init_db(config_dir):
    global Db
    db_path = os.path.join(config_dir, "vpn-sandbox.db")
    Db = sqlite3.connect(db_path, check_same_thread=False)
    
    cursor = Db.cursor()
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS configs (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL UNIQUE,
            config JSON NOT NULL
        )
    """)
    Db.commit()

def save_config(name, config):
    config_str = json.dumps(config)
    cursor = Db.cursor()
    cursor.execute("""
        INSERT OR REPLACE INTO configs (name, config)
        VALUES (?, ?)
    """, (name, config_str))
    Db.commit()

def get_config(name):
    cursor = Db.cursor()
    cursor.execute("SELECT config FROM configs WHERE name = ?", (name,))
    row = cursor.fetchone()
    if row is None:
        return None, Exception("not found")
    try:
        config = json.loads(row[0])
        return config, None
    except Exception as e:
        return None, e
